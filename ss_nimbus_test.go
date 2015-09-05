package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"io/ioutil"
)

var us_blocks = []string{
	"00000000544a80ff5b70e836099a4b9b431af932bb35a7dffb4544e9e0e8901f",
	"00000000c0374dbda3681ec59a05ccf12fd1e3e7b7893a319872f89b081607fb",
	"000000014cdb111d944d4bfbcc1436762121c9114d369ac3d6c52b7e2ac0d680",
	"000000017d4bab3810b07dd49eaad78dcfc293f6f176a2d023cc910ce2ed2e55",
	"000000033a63b2c61e09aba72c6f12b1fda379fcb474ad020695bcd51dee240b",
}

var eu_blocks = []string{
	"0000000c3a9204897d803c93c0d780f24ef2c58791cc83f760663670bb27e221",
	"0000000d37eb0c5cf0d4a023d4c6291177b2cdce20e6dfda3e18642b1b28f78b",
	"0000001199ca1c100eba6eeaa819033cd5e370894761eca937aaa31c67adbc4e",
	"000000154a114f5225919781ef0b1316dcd7bd2672444fb557ffa83c06919918",
	"0000001ad2e53aee89fadaae69dce004ce6614dc37854fb664f849122a7406c4",
}

var ap_blocks = []string{
	"0000000000000000000000000000000000000000000000000000000000000000",
	"00000000c0374dbda3681ec59a05ccf12fd1e3e7b7893a319872f89b081607fb",
	"000000019c0b83ac221e8c342fd81297d6d1c7cb8de79882914f7529df75fd95",
	"00000001b0571fc5648a7b59654ba2ad6aea34ff0fc2f02b657374a84ccfc6fa",
	"00000001b77332705b025290d08410693c650b03e2d536549d4b88891f7b64c1",
}

var mix_region_blocks = []string{
	"00000001b0571fc5648a7b59654ba2ad6aea34ff0fc2f02b657374a84ccfc6fa", // ap
	"00000001b77332705b025290d08410693c650b03e2d536549d4b88891f7b64c1", // ap
	"0000000d37eb0c5cf0d4a023d4c6291177b2cdce20e6dfda3e18642b1b28f78b", // eu
	"0000001199ca1c100eba6eeaa819033cd5e370894761eca937aaa31c67adbc4e", // eu
	"00000000544a80ff5b70e836099a4b9b431af932bb35a7dffb4544e9e0e8901f", // us
}

var mix_result_blocks = []string{
	"00000001b0571fc5648a7b59654ba2ad6aea34ff0fc2f02b657374a8ffffffff", // ap false
	"00000001b77332705b025290d08410693c650b03e2d536549d4b8889ffffffff", // ap false
	"0000000d37eb0c5cf0d4a023d4c6291177b2cdce20e6dfda3e18642b1b28f78b", // eu
	"0000001199ca1c100eba6eeaa819033cd5e370894761eca937aaa31c67adbc4e", // eu
	"00000000544a80ff5b70e836099a4b9b431af932bb35a7dffb4544e9ffffffff", // us false
}

type Test struct {
	blocks []string
	expected int
	checkids []string
}

var BlockTests = map[string]*Test{
	"us": &Test{
		blocks: us_blocks,
		expected: http.StatusOK,
		checkids: []string{},
	},
	"eu": &Test{
		blocks: eu_blocks,
		expected: http.StatusOK,
		checkids: []string{},
	},
	"ap": &Test{
		blocks: ap_blocks,
		expected: http.StatusOK,
		checkids: []string{},
	},
	"mr": &Test{
		blocks: mix_region_blocks,
		expected: http.StatusOK,
		checkids: []string{},
	},
	"me": &Test{
		blocks: mix_result_blocks,
		expected: http.StatusNotFound,
		checkids: []string{
		"00000001b0571fc5648a7b59654ba2ad6aea34ff0fc2f02b657374a8ffffffff", // ap false
		"00000001b77332705b025290d08410693c650b03e2d536549d4b8889ffffffff", // ap false
		"00000000544a80ff5b70e836099a4b9b431af932bb35a7dffb4544e9ffffffff", // us false
		},
	},
}

func Test_CheckBlock(t *testing.T) {

	config := "config.ini"
	handle_config(&config)
	FillPool(100)

	for k, v := range BlockTests {
		b, err := json.Marshal(v.blocks)
		if err != nil {
			t.Error(err.Error())
		}

		body := bytes.NewBuffer(b)

		req, _ := http.NewRequest("POST", "/blocks", body)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != v.expected {
			t.Errorf("%s check didn't return %v", k, v.expected)
		}

		if w.Code == http.StatusNotFound {
			rbody, _ := ioutil.ReadAll(w.Body)
			missingIds := make([]string, 0, 100)
			json.Unmarshal(rbody, &missingIds)
			for _, id := range missingIds{
				found := false
				for _, idtc := range v.checkids {
					if id == idtc {
						found = true
						break
					} else {
						continue
					}
				}
				if !found {
					t.Errorf("id %s was not an expected id to find", id)
				}
			}
		}
	}
}