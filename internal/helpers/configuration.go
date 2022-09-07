package helpers

import (
	"encoding/json"
	"flag"
	"github.com/be-ys-cloud/antares/internal/structures"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
)

var Configuration structures.Configuration

func init() {
	configFile := ""

	logrus.Infoln("Reading informations from command line.")

	// Retrieve arguments from command line ( + defaults & help )
	flag.StringVar(&Configuration.KeePassFile, "keepassFile", "vault-keepass.kdbx", "Override default output file (default: data.kbdx)")
	flag.StringVar(&configFile, "configFile", "configuration.json", "Override configuration file to read (default: configuration.json)")
	flag.StringVar(&Configuration.MasterPassword, "password", "", "Password of the generated kbdx file. (mandatory)")
	flag.StringVar(&Configuration.VaultAddress, "vaultServer", "", "Vault address (mandatory)")
	flag.StringVar(&Configuration.VaultToken, "vaultToken", "", "Vault token (mandatory)")
	flag.BoolVar(&Configuration.Export, "export", false, "Import a Keepass into Vault")
	flag.BoolVar(&Configuration.Import, "import", false, "Export Vault data to a Keepass")
	flag.Parse()

	if Configuration.VaultToken == "" {
		logrus.Infoln("Could not read vault token from CLI. Trying from environment...")
		Configuration.VaultToken = os.Getenv("ANTARES_VAULT_TOKEN")
	}

	if Configuration.MasterPassword == "" {
		logrus.Infoln("Could not read keepass master from CLI. Trying from environment...")
		Configuration.MasterPassword = os.Getenv("ANTARES_KEEPASS_PASSWORD")
	}

	if (Configuration.Export && Configuration.Import) || (!Configuration.Export && !Configuration.Import) {
		logrus.Fatalln("You must specify one (and only one) option between -import and -export.")
	}

	if Configuration.KeePassFile == "" || Configuration.MasterPassword == "" || Configuration.VaultAddress == "" || Configuration.VaultToken == "" {
		logrus.Fatalln("One of the mandatory variables are not defined.")
	}

	if !strings.HasSuffix(Configuration.VaultAddress, "/") {
		Configuration.VaultAddress += "/"
	}

	// If we are exporting secrets, we must load configuration file and get paths to be exported
	if Configuration.Export {
		if configFile == "" {
			logrus.Fatalln("One of the mandatory variables are not defined.")
		}

		// Load configuration

		file, err := os.Open(configFile)
		if err != nil {
			logrus.Fatalf("Unable to load configuration file ! (Filename : %s)", configFile)
		}

		fileContent, _ := ioutil.ReadAll(file)
		_ = json.Unmarshal(fileContent, &Configuration.PathsToExport)
		_ = file.Close()

		logrus.Infoln("Configuration loaded.")
	}
}
