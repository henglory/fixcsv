package fixcsv

import "testing"

type testInfo struct {
	Test1 string  `fixcsv:"1:5"`
	Test2 string  `fixcsv:"2:8"`
	Test3 string  `fixcsv:"3:8"`
	Test4 string  `fixcsv:"4:5"`
	Test5 string  `fixcsv:"5:5"`
	Test6 float64 `fixcsv:"6:10"`
}

func TestMarshaling(t *testing.T) {
	info := testInfo{

		Test1: "123456789",
		Test2: "123",
		Test3: "22222222222",
		Test4: "123123123",
		Test5: "0",
		Test6: 12.0,
	}
	b, err := Marshal(info)
	if err != nil {
		t.Error(err)
	}
	bstr := string(b)
	if bstr != "12345||123||22222222||12312||0||12" {
		t.Errorf("result is wrong %s", bstr)
	}
}
