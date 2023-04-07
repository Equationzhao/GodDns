package Device

import (
	"GodDns/Core"
	"strconv"
	"strings"
	"testing"
)

func TestDevice_GenerateConfigInfo(t *testing.T) {
	d := Device{
		Devices: []string{"eth0", "eth1"},
	}
	info, err := d.GenerateConfigInfo(d, 0)
	if err != nil {
		t.Error(err)

	}
	t.Log(info.Content)
}

func TestDevice_ReadConfig(t *testing.T) {
	d := Device{
		Devices: []string{"eth0", "eth1"},
	}

	t.Log(d)

	parameters, err, errs := Core.ConfigureReader("test.ini", ConfigFactoryInstance)

	if errs != nil {
		t.Error(errs)
	}

	if err != nil {
		t.Error(err)
	} else if len(parameters) != 1 {
		t.Error("wrong number of parameters")
	}

	if d_, ok := parameters[0].(Device); ok {
		t.Log(d_)
		for i, ds1 := range d.GetDevices() {
			if ds1 != d_.GetDevices()[i] {
				t.Error("wrong device name")
			}
		}
	} else {
		t.Error("wrong type ", d)
	}

}

func TestConvert2StringSlice(t *testing.T) {
	deviceList := "[eth0,eth1]"
	s := strings.Split(strings.Trim(deviceList, "[]"), ",")

	if len(s) != 2 {
		t.Error("wrong length")
	}

	for i, d := range s {
		if d != "eth"+strconv.Itoa(i) {
			t.Errorf("wrong device name %s!=%s", d, "eth"+strconv.Itoa(i+1))
		} else {
			t.Log(d)
		}
	}

}
