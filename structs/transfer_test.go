package structs

import (
	"fmt"
	"testing"

	"github.com/mailru/easyjson"
)

func TestUnmarshalInfo(t *testing.T) {
	info := Info{}
	bs, err := easyjson.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	var newInfo Info
	err = easyjson.Unmarshal(bs, &newInfo)
	if err != nil {
		t.Fatal(err)
	}
	if info != newInfo {
		t.Errorf("info != newInfo; info: %v, newInfo: %v",
			info, newInfo)
	}
}

func TestUnmarshalURLObject(t *testing.T) {
	urlObj := URLObject{}
	bs, err := easyjson.Marshal(urlObj)
	if err != nil {
		t.Fatal(err)
	}
	var newURLObj URLObject
	err = easyjson.Unmarshal(bs, &newURLObj)
	if err != nil {
		t.Fatal(err)
	}
	if urlObj != newURLObj {
		t.Errorf("urlObj != newURLObj; urlObj: %v, newURLObj: %v",
			urlObj, newURLObj)
	}
}

func TestUnmarshalSearchObject(t *testing.T) {
	searchObj := SearchObject{}
	bs, err := easyjson.Marshal(searchObj)
	if err != nil {
		t.Fatal(err)
	}
	var newSearchObj SearchObject
	err = easyjson.Unmarshal(bs, &newSearchObj)
	if err != nil {
		t.Fatal(err)
	}
	if searchObj != newSearchObj {
		t.Errorf("searchObj != newSearchObj ; searchObj: %v, newSearchObj: %v",
			searchObj, newSearchObj)
	}
}

func TestUnmarshalSearchObjectNotNilAll(t *testing.T) {
	globalID := 3
	systemObjectID := "90"
	id := 42
	idEn := 43
	mode := ""
	modeEn := "a"
	searchObj := SearchObject{
		GlobalID:       &globalID,
		SystemObjectID: &systemObjectID,
		ID:             &id,
		IDEn:           &idEn,
		Mode:           &mode,
		ModeEn:         &modeEn,
	}
	bs, err := easyjson.Marshal(searchObj)
	if err != nil {
		t.Fatal(err)
	}
	var newSearchObj SearchObject
	err = easyjson.Unmarshal(bs, &newSearchObj)
	if err != nil {
		t.Fatal(err)
	}
	if *searchObj.SystemObjectID != *newSearchObj.SystemObjectID {
		t.Errorf("searchObj != newSearchObj ; searchObj: %v, newSearchObj: %v",
			searchObj, newSearchObj)
	}
	if *searchObj.GlobalID != *newSearchObj.GlobalID {
		t.Errorf("searchObj != newSearchObj ; searchObj: %v, newSearchObj: %v",
			searchObj, newSearchObj)
	}
	if *searchObj.ID != *newSearchObj.ID {
		t.Errorf("searchObj != newSearchObj ; searchObj: %v, newSearchObj: %v",
			searchObj, newSearchObj)
	}
	if *searchObj.IDEn != *newSearchObj.IDEn {
		t.Errorf("searchObj != newSearchObj ; searchObj: %v, newSearchObj: %v",
			searchObj, newSearchObj)
	}
	if *searchObj.Mode != *newSearchObj.Mode {
		t.Errorf("searchObj != newSearchObj ; searchObj: %v, newSearchObj: %v",
			searchObj, newSearchObj)
	}
	if *searchObj.ModeEn != *newSearchObj.ModeEn {
		t.Errorf("searchObj != newSearchObj ; searchObj: %v, newSearchObj: %v",
			searchObj, newSearchObj)
	}
	if searchObj.Offset != newSearchObj.Offset {
		t.Errorf("searchObj != newSearchObj ; searchObj: %v, newSearchObj: %v",
			searchObj, newSearchObj)
	}
}

func TestUnmarshalPaginationObject(t *testing.T) {
	paginationObject := PaginationObject{}
	bs, err := easyjson.Marshal(paginationObject)
	if err != nil {
		t.Fatal(err)
	}
	var newPaginationObject PaginationObject
	err = easyjson.Unmarshal(bs, &newPaginationObject)
	if err != nil {
		t.Fatal(err)
	}
	errString := fmt.Sprintf("paginationObject != newPaginationObject; paginationObject: %v, newPaginationObject: %v",
		paginationObject, newPaginationObject)
	if paginationObject.HasPrevious != newPaginationObject.HasPrevious {
		t.Errorf(errString)
	}
	if paginationObject.HasNext != newPaginationObject.HasNext {
		t.Errorf(errString)
	}
	if paginationObject.Size != newPaginationObject.Size {
		t.Errorf(errString)
	}
	if paginationObject.Offset != newPaginationObject.Offset {
		t.Errorf(errString)
	}
	if len(paginationObject.Data) != len(newPaginationObject.Data) {
		t.Errorf(errString)
	}
}
