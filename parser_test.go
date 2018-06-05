package drushmakeparser

import (
	"sort"
	"testing"
)

var testBlocks = []struct {
	in      string
	out     Package
	success bool
}{
	{
		in: `projects[ns_core][type] = module
projects[ns_core][download][type] = git
projects[ns_core][download][branch] = 7.x-2.x`,
		out: Package{
			Name:     "ns_core",
			Type:     "module",
			Download: Download{Type: "git", Branch: "7.x-2.x"},
		},
		success: true,
	},
	{
		in: `projects[draggableviews][type] = module
projects[draggableviews][download][type] = git
projects[draggableviews][download][revision] = 9677bc18b7255e13c33ac3cca48732b855c6817d
projects[draggableviews][download][branch] = 7.x-2.x`,
		out: Package{
			Name:     "draggableviews",
			Type:     "module",
			Download: Download{Type: "git", Revision: "9677bc18b7255e13c33ac3cca48732b855c6817d", Branch: "7.x-2.x"},
		},
		success: true,
	},
	{
		in: `projects[views] = 3.1`,
		out: Package{
			Name:    "views",
			Version: "3.1",
		},
		success: true,
	},
	{
		in: `projects[nodequeue][subdir] = contrib
projects[nodequeue][version] = 2.0-alpha1
projects[nodequeue][patch][] = "http://drupal.org/files/issues/1023606-qid-to-name-6.patch"
projects[nodequeue][patch][] = "http://drupal.org/files/issues/nodequeue_d7_autocomplete-872444-6.patch"`,
		out: Package{
			Name:    "nodequeue",
			Version: "2.0-alpha1",
			Patch: []string{
				"http://drupal.org/files/issues/1023606-qid-to-name-6.patch",
				"http://drupal.org/files/issues/nodequeue_d7_autocomplete-872444-6.patch",
			},
		},
		success: true,
	},
}

var testMakefile string = `projects[media] = 2.x-dev
projects[media_youtube][version] = 1.0-alpha5
projects[media_youtube][subdir] = media_plugins
projects[media_flickr][version] = 1.0-alpha1
projects[media_flickr][subdir] = media_plugins
projects[rubik] = 4.0-beta7
projects[rubik][patch][] = "http://drupal.org/files/rubik-print-css.patch"
projects[nodequeue][subdir] = contrib
projects[nodequeue][version] = 2.0-alpha1
projects[nodequeue][patch][] = "http://drupal.org/files/issues/1023606-qid-to-name-6.patch"
projects[nodequeue][patch][] = "http://drupal.org/files/issues/nodequeue_d7_autocomplete-872444-6.patch"`

var testFullComponentList = struct {
	in  string
	out DrushMakeInfo
}{
	in: testMakefile,
	out: DrushMakeInfo{
		[]Package{
			{Name: "media", Version: "2.x-dev"},
			{Name: "media_youtube", Version: "1.0-alpha2"},
			{Name: "media_flickr", Version: "1.0-alpha1"},
			{Name: "rubik", Version: "4.0-beta7", Patch: []string{"http://drupal.org/files/rubik-print-css.patch"}},
			{Name: "nodequeue", Version: "2.0-alpha1",
				Patch: []string{
					"http://drupal.org/files/issues/1023606-qid-to-name-6.patch",
					"http://drupal.org/files/issues/nodequeue_d7_autocomplete-872444-6.patch",
				},
			},
		},
	},
}

// Test the processing of the complete string block.
func TestProjectNameListParser(t *testing.T) {
	var testComponentList = struct {
		in  string
		out []string
	}{
		in:  testMakefile,
		out: []string{"media", "media_youtube", "media_flickr", "rubik", "nodequeue"},
	}

	if components := componentList(testComponentList.in); !LazyArrayEqual(components, testComponentList.out) {
		t.Error("For", testComponentList.in, "expected", testComponentList.out, "got", components)
	}
}

// Test if component blocks are correctly parsed and populated.
func TestBlockToComponentParser(t *testing.T) {
	for _, testBlock := range testBlocks {
		// First pass the snippet to the block parser.
		testPackage := Package{}
		testPackage.Parse(testBlock.in)

		if success := (testPackage.Name == testBlock.out.Name); !success && success != testBlock.success {
			t.Error("For", testBlock.in, "expected", testBlock.out.Name, "got", testPackage.Name)
		}
		if success := (testPackage.Type == testBlock.out.Type); !success && success != testBlock.success {
			t.Error("For", testBlock.in, "expected", testBlock.out.Type, "got", testPackage.Type)
		}
		if success := (testPackage.Version == testBlock.out.Version); !success && success != testBlock.success {
			t.Error("For", testBlock.in, "expected", testBlock.out.Version, "got", testPackage.Version)
		}
		if success := (testPackage.Download == testBlock.out.Download); !success && success != testBlock.success {
			t.Error("For", testBlock.in, "expected", testBlock.out.Download, "got", testPackage.Download)
		}
		if success := LazyArrayEqual(testPackage.Patch, testBlock.out.Patch); !success && success != testBlock.success {
			t.Error("For", testBlock.in, "expected", testBlock.out.Patch, "got", testPackage.Patch)
		}
	}
}

// Helper function to compare arrays regardless of the order of values.
func LazyArrayEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Sort(sort.StringSlice(a))
	sort.Sort(sort.StringSlice(b))

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
