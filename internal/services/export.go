package services

import (
	"crypto/tls"
	"encoding/json"
	"github.com/be-ys-cloud/antares/internal/helpers"
	"github.com/be-ys-cloud/antares/internal/structures"
	"github.com/sirupsen/logrus"
	"github.com/tobischo/gokeepasslib/v3"
	"github.com/tobischo/gokeepasslib/v3/wrappers"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func Export() {
	logrus.Infoln("Starting listing of secrets. This step may take a long time if you have a lot of secrets, or a wide directory.")
	for i, pathToExport := range helpers.Configuration.PathsToExport {
		for _, Path := range pathToExport.Paths {
			logrus.Infof("Starting listing of secrets in %s", pathToExport.Engine+Path)
			helpers.Configuration.PathsToExport[i].FullPaths = append(pathToExport.FullPaths, parseEnginePaths(Path, pathToExport.Engine)...)
			logrus.Infof("Listing finished for secrets in %s", pathToExport.Engine+Path)
		}
	}

	logrus.Infoln("Listing finished. Starting exportation.")
	createFileAndGetSecrets()
	logrus.Infoln("Exportation complete. You can now open your file with KeePass2 or a similar tool.")
}

// ----- Private methods
func createFileAndGetSecrets() {

	file, err := os.Create(helpers.Configuration.KeePassFile)
	if err != nil {
		logrus.Fatalf("Could not create file %s. Error was : %s", helpers.Configuration.KeePassFile, err.Error())
	}
	defer file.Close()

	rootGroup := gokeepasslib.NewGroup()
	rootGroup.Name = "Secrets"

	for _, key := range helpers.Configuration.PathsToExport {
		subGroup := gokeepasslib.NewGroup()
		subGroup.Name = key.Engine

		for _, key2 := range key.FullPaths {
			if field, err := getDataFromServer[structures.DataList](helpers.Configuration.VaultAddress+"v1/"+key.Engine+"/data"+key2, "GET", helpers.Configuration.VaultToken); err == nil {
				subGroup = parseAndAddGroup(subGroup, key2[1:], field)
			}
		}

		rootGroup.Groups = append(rootGroup.Groups, subGroup)

	}

	// now create the database containing the root group
	db := &gokeepasslib.Database{
		Header:      gokeepasslib.NewHeader(),
		Credentials: gokeepasslib.NewPasswordCredentials(helpers.Configuration.MasterPassword),
		Content: &gokeepasslib.DBContent{
			Meta: gokeepasslib.NewMetaData(),
			Root: &gokeepasslib.RootData{
				Groups: []gokeepasslib.Group{rootGroup},
			},
		},
	}

	// Lock entries using stream cipher
	_ = db.LockProtectedEntries()

	// and encode it into the file
	keepassEncoder := gokeepasslib.NewEncoder(file)
	if err := keepassEncoder.Encode(db); err != nil {
		logrus.Fatalf("A fatal error occured : %s", err.Error())
	}

}

func parseAndAddGroup(group gokeepasslib.Group, path string, dataContent structures.DataList) gokeepasslib.Group {
	subGroup := gokeepasslib.NewGroup()

	if strings.Contains(path, "/") {

		subGroup.Name = strings.Split(path, "/")[0]

		path = strings.TrimPrefix(path, subGroup.Name+"/")
		found := false

		for i := range group.Groups {
			if group.Groups[i].Name == subGroup.Name {
				path = strings.TrimPrefix(path, group.Groups[i].Name+"/")
				group.Groups[i] = parseAndAddGroup(group.Groups[i], path, dataContent)
				found = true
				continue
			}
		}
		if !found {
			group.Groups = append(group.Groups, parseAndAddGroup(subGroup, path, dataContent))
		}

	} else {
		if group.Name == path {
			group = createEntriesAndAddToGroup(dataContent, group)
		} else {
			subGroup.Name = path
			subGroup = createEntriesAndAddToGroup(dataContent, subGroup)
			group.Groups = append(group.Groups, subGroup)
		}
	}

	return group
}

func createEntriesAndAddToGroup(dataContent structures.DataList, group gokeepasslib.Group) gokeepasslib.Group {
	for itemName, itemValue := range dataContent.Data.Data {
		subEntry := gokeepasslib.NewEntry()
		subEntry.Values = append(subEntry.Values, mkValue("Title", itemName, false))
		subEntry.Values = append(subEntry.Values, mkValue("UserName", itemName, false))
		subEntry.Values = append(subEntry.Values, mkValue("Password", itemValue, true))
		group.Entries = append(group.Entries, subEntry)
	}
	return group
}

func parseEnginePaths(path string, engine string) (returnList []string) {
	if !strings.HasSuffix(path, "/") {
		return []string{path}
	}

	var field structures.MetadataList
	field, err := getDataFromServer[structures.MetadataList](helpers.Configuration.VaultAddress+"v1/"+engine+"/metadata"+path, "LIST", helpers.Configuration.VaultToken)

	if err == nil {
		for _, k := range field.Data.Keys {
			if strings.HasSuffix(k, "/") {
				list := parseEnginePaths(path+k, engine)
				for _, l := range list {
					returnList = append(returnList, l)
				}
			} else {
				returnList = append(returnList, path+k)
			}
		}
	}
	return
}

func getDataFromServer[T any](url string, method string, token string) (data T, err error) {
	res, _ := http.NewRequest(method, url, nil)
	res.Header.Set("X-Vault-Token", token)

	client := &http.Client{Timeout: time.Second * 20, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	resp, err := client.Do(res)

	if err != nil {
		logrus.Warningf("Unable to get path %s. Internal script error.", url)
		return
	}

	if resp.StatusCode != 200 {
		logrus.Warningf("Unable to get path %s. Server returned code %d", url, resp.StatusCode)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	_ = json.Unmarshal(body, &data)

	return
}

func mkValue(key string, value string, protected bool) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value, Protected: wrappers.BoolWrapper{Bool: protected}}}
}
