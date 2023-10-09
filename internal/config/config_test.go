package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	assert := assert.New(t)

	configVars["bela"] = configVar{
		Type:    boolType,
		Default: true,
		EnvVar:  "BELA",
	}
	configVars["belastr"] = configVar{
		Type:    stringType,
		Default: "belabacsi",
		EnvVar:  "BELA_BACSI",
	}
	configVars["belatooverridestr"] = configVar{
		Type:    stringType,
		Default: "belabacsi",
		EnvVar:  "BELA_BACSI",
	}
	configVars["notdefined.susu"] = configVar{
		Type:    stringType,
		Default: "korte",
		EnvVar:  "NOTDEFINED_SUSU",
	}
	configVars["notdefined.kiralyfi"] = configVar{
		Type:    stringType,
		Default: "karomeros",
		EnvVar:  "NOTDEFINED_KIRALYFI",
	}
	configFile = "testdata/config-test.yaml"

	conf, err := NewConfig()
	assert.NotNil(conf)
	assert.Nil(err)

	assert.Equal(true, conf.GetBool("bela"))
	assert.Equal("belabacsi", conf.GetString("belastr"))
	assert.Equal("belabacsi_overrided", conf.GetString("belatooverridestr"))
	assert.Equal("sub", conf.GetString("notdefined.subitem"))
	assert.Equal("korte", conf.GetString("notdefined.susu"))
	assert.Equal("szivemvidam", conf.GetString("notdefined.kiralyfi"))

	os.Setenv("BELA", "false")
	os.Setenv("BELA_BACSI", "susu")
	assert.Equal(false, conf.GetBool("bela"))
	assert.Equal("susu", conf.GetString("belastr"))
	os.Setenv("NOTDEFINED_KIRALYFI", "kiralyfi")
	assert.Equal("kiralyfi", conf.GetString("notdefined.kiralyfi"))

	standuppers := conf.GetStringSlice("standuppers")
	assert.Len(standuppers, 3)
	assert.Equal(standuppers[0], "TestBela")
	os.Setenv("BELA_BACSI", "susu")
}
