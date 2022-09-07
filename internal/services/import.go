package services

import (
	"crypto/tls"
	"encoding/json"
	"github.com/be-ys-cloud/antares/internal/helpers"
	"github.com/sirupsen/logrus"
	"github.com/tobischo/gokeepasslib/v3"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var success = 0
var errors = 0

func Import() {
	file, _ := os.Open(helpers.Configuration.KeePassFile)

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(helpers.Configuration.MasterPassword)
	_ = gokeepasslib.NewDecoder(file).Decode(db)

	err := db.UnlockProtectedEntries()
	if err != nil {
		log.Fatalln(err.Error())
	}

	for _, k := range db.Content.Root.Groups[0].Groups {
		for _, j := range k.Groups {
			sendSecretToVault(j, "", k.Name)
		}
	}

	logrus.Infof("Importation done. %d secrets imported, %d secrets failed.", success, errors)
}

func sendSecretToVault(g gokeepasslib.Group, path string, engine string) {
	for _, j := range g.Groups {
		sendSecretToVault(j, path+g.Name+"/", engine)
	}
	if len(g.Entries) != 0 {

		data := map[string]string{}
		for _, j := range g.Entries {
			data[j.GetTitle()] = j.GetPassword()
		}

		data2 := map[string]interface{}{}
		data2["data"] = data

		d, _ := json.Marshal(data2)

		readyToSend := strings.NewReader(string(d))

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		// Send request to Grafana Server
		addr := helpers.Configuration.VaultAddress + "v1/" + engine + "/data/" + path + g.Name
		logrus.Infof("Creating secret %s", addr)
		res, _ := http.NewRequest("POST", addr, readyToSend)
		res.Header.Set("X-Vault-Token", helpers.Configuration.VaultToken)
		res.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: time.Second * 20, Transport: tr}
		resp, err := client.Do(res)

		if err != nil || resp.StatusCode != 200 {
			logrus.Warningf("Unable to create secret %s", addr)
			errors += 1
		} else {
			success += 1
			logrus.Infof("Secret %s created.", addr)
		}
	}
}
