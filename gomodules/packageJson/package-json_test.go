package packageJson

import (
	"testing"
)

func TestReadPackageJson(t *testing.T) {

	equalInterface := func(a map[string]interface{}, b map[string]string) bool {
		for k, v := range a {
			if b[k] != v.(string) {
				return false
			}
		}
		return true
	}

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		assert  func(*PackageJSON) bool
		wantErr bool
	}{
		{"read author as string", args{path: "./testdata/authorAsString.json"}, func(p *PackageJSON) bool {
			return p.Author == "Barney Rubble <b@rubble.com> (http://barnyrubble.tumblr.com/)"
		}, false},
		{"read author as Objet", args{path: "./testdata/authorAsObject.json"}, func(p *PackageJSON) bool {
			return equalInterface(p.Author.(map[string]interface{}), map[string]string{
				"name":  "Barney Rubble",
				"email": "b@rubble.com",
				"url":   "http://barnyrubble.tumblr.com/",
			})
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Read(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadPackageJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.assert(got) {
				t.Errorf("ReadPackageJson():  = %v", got)
			}
		})
	}
}
