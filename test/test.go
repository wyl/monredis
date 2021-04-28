package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

func extractData(srcField string, data map[string]interface{}) (result interface{}, err error) {
	//result := make(map[string]interface{}, 0)
	var cur = data
	fields := strings.Split(srcField, ".")
	for i, field := range fields {
		if i+1 == len(fields) {
			result = cur[field]
		} else {
			var tmpArry []interface{}
			if next, ok := cur[field].([]interface{}); ok {
				for _, n := range next {
					if nns, ok := n.(map[string]interface{}); ok {
						v, _ := extractData(strings.Join(fields[i+1:], "."), nns)
						tmpArry = append(tmpArry, v)
					}
				}
				result = tmpArry
				break
			} else if next, ok := cur[field].(map[string]interface{}); ok {
				cur = next
			} else {
				break
			}
		}
	}
	if result == nil {
		var detail interface{}
		b, e := json.Marshal(data)
		if e == nil {
			detail = string(b)
		} else {
			detail = err
		}
		err = fmt.Errorf("Source field %s not found in document: %s", srcField, detail)
	}
	return result, nil
}

func main() {
	d := `{"_id":"01bfc940200a11ebb7f2717560a5a3b6","c":{"id":1},"cid":["11","22"],"connections":[{"id":"94561100e02d11ea95a457cc85321374","props":{},"type":"entity"},{"id":"94498de0e02d11ea95a457cc85321374","props":{},"type":"entity"},{"id":"01531de0e91911ea95a457cc85321374","props":{},"type":"entity"},{"id":"86a03a30e02911ea867a3b2c54e4df9b","props":{},"type":"entity"},{"id":"95125c20e02d11ea95a457cc85321374","props":{},"type":"entity"}],"created_at":1604651291604,"status":1,"tag":"电影2","title":"喜羊羊大电影","type":"series","updated":"2021-01-14T06:23:26.862Z","updated_at":1609328795357}`
	var data map[string]interface{}
	json.Unmarshal([]byte(d), &data)
	kk, _ := extractData("connections.id", data)
	fmt.Println("output", kk)

}
