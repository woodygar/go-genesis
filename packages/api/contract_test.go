// MIT License
//
// Copyright (c) 2016 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package api

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/GenesisCommunity/go-genesis/packages/crypto"
)

func TestNewContracts(t *testing.T) {

	wanted := func(name, want string) bool {
		var ret getTestResult
		err := sendPost(`test/`+name, nil, &ret)
		if err != nil {
			t.Error(err)
			return false
		}
		if ret.Value != want {
			t.Error(fmt.Errorf(`%s != %s`, ret.Value, want))
			return false
		}
		return true
	}

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	for _, item := range contracts {
		var ret getContractResult
		err := sendGet(`contract/`+item.Name, nil, &ret)
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(apiErrors[`E_CONTRACT`], item.Name)) {
				form := url.Values{"Name": {item.Name}, "Value": {item.Value},
					"Conditions": {`true`}}
				if err := postTx(`NewContract`, &form); err != nil {
					if item.Params[0].Results[`error`] != err.Error() {
						t.Error(err)
						return
					}
					continue
				}
			} else {
				t.Error(err)
				return
			}
		}
		if strings.HasSuffix(item.Name, `testUpd`) {
			continue
		}
		for _, par := range item.Params {
			form := url.Values{}
			for key, value := range par.Params {
				form[key] = []string{value}
			}
			if err := postTx(item.Name, &form); err != nil {
				t.Error(err)
				return
			}
			for key, value := range par.Results {
				if !wanted(key, value) {
					return
				}
			}
		}
	}
}

var contracts = []smartContract{
	{`errTestMessage`, `contract errTestMessage {
		conditions {
		}
		action { qvar ivar int}
	}`,
		[]smartParams{
			{nil, map[string]string{`error`: `{"type":"panic","error":"unknown variable qvar"}`}},
		}},

	{`EditProfile9`, `contract EditProfile9 {
		data {
		}
		conditions {
		}
		action {
			var ar array
			ar = Split("point 1,point 2", ",")
			Test("split",  Str(ar[1]))
			$ret = DBFind("contracts").Columns("id,value").Where("id>= ? and id<= ?",3,5).Order("id")
			Test("edit",  "edit value 0")
		}
	}`,
		[]smartParams{
			{nil, map[string]string{`edit`: `edit value 0`, `split`: `point 2`}},
		}},

	{`TestDBFindOK`, `
		contract TestDBFindOK {
		action {
			var ret array
			var vals map
			ret = DBFind("contracts").Columns("id,value").Where("id>= ? and id<= ?",3,5).Order("id")
			if Len(ret) {
				Test("0",  "1")	
			} else {
				Test("0",  "0")	
			}
			ret = DBFind("contracts").Limit(3)
			if Len(ret) == 3 {
				Test("1",  "1")	
			} else {
				Test("1",  "0")	
			}
			ret = DBFind("contracts").Order("id").Offset(1).Limit(1)
			if Len(ret) != 1 {
				Test("2",  "0")	
			} else {
				vals = ret[0]
				Test("2",  vals["id"])	
			}
			ret = DBFind("contracts").Columns("id").Order("id").Offset(1).Limit(1)
			if Len(ret) != 1 {
				Test("3",  "0")	
			} else {
				vals = ret[0]
				Test("3", vals["value"] + vals["id"])	
			}
			ret = DBFind("contracts").Columns("id").Where("id='1'")
			if Len(ret) != 1 {
				Test("4",  "0")	
			} else {
				vals = ret[0]
				Test("4", vals["id"])	
			}
			ret = DBFind("contracts").Columns("id").Where("id='1'")
			if Len(ret) != 1 {
				Test("4",  "0")	
			} else {
				vals = ret[0]
				Test("4", vals["id"])	
			}
			ret = DBFind("contracts").Columns("id,value").Where("id> ? and id < ?", 3, 8).Order("id")
			if Len(ret) != 4 {
				Test("5",  "0")	
			} else {
				vals = ret[0]
				Test("5", vals["id"])	
			}
			ret = DBFind("contracts").WhereId(7)
			if Len(ret) != 1 {
				Test("6",  "0")	
			} else {
				vals = ret[0]
				Test("6", vals["id"])	
			}
			var one string
			one = DBFind("contracts").WhereId(5).One("id")
			Test("7",  one)	
			var row map
			row = DBFind("contracts").WhereId(3).Row()
			Test("8",  row["id"])	
			Test("255",  "255")	
		}
	}`,
		[]smartParams{
			{nil, map[string]string{`0`: `1`, `1`: `1`, `2`: `2`, `3`: `2`, `4`: `1`, `5`: `4`,
				`6`: `7`, `7`: `5`, `8`: `3`, `255`: `255`}},
		}},
	{`testEmpty`, `contract testEmpty {
				action { Test("empty",  "empty value")}}`,
		[]smartParams{
			{nil, map[string]string{`empty`: `empty value`}},
		}},
	{`testUpd`, `contract testUpd {
					action { Test("date",  "-2006.01.02-")}}`,
		[]smartParams{
			{nil, map[string]string{`date`: `-` + time.Now().Format(`2006.01.02`) + `-`}},
		}},
	{`testLong`, `contract testLong {
		action { Test("long",  "long result")
			$result = DBFind("contracts").WhereId(2).One("value") + DBFind("contracts").WhereId(4).One("value")
			Println("Result", $result)
			Test("long",  "long result")
		}}`,
		[]smartParams{
			{nil, map[string]string{`long`: `long result`}},
		}},
	{`testSimple`, `contract testSimple {
				data {
					amount int
					name   string
				}
				conditions {
					Test("scond", $amount, $name)
				}
				action { Test("sact", $name, $amount)}}`,
		[]smartParams{
			{map[string]string{`name`: `Simple name`, `amount`: `-56781`},
				map[string]string{`scond`: `-56781Simple name`,
					`sact`: `Simple name-56781`}},
		}},
	{`errTestVar`, `contract errTestVar {
			conditions {
			}
			action { var ivar int}
		}`,
		nil},
}

func TestEditContracts(t *testing.T) {

	wanted := func(name, want string) bool {
		var ret getTestResult
		err := sendPost(`test/`+name, nil, &ret)
		if err != nil {
			t.Error(err)
			return false
		}
		if ret.Value != want {
			t.Error(fmt.Errorf(`%s != %s`, ret.Value, want))
			return false
		}
		return true
	}

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var cntlist contractsResult
	err := sendGet(`contracts`, nil, &cntlist)
	if err != nil {
		t.Error(err)
		return
	}
	var ret getContractResult
	err = sendGet(`contract/testUpd`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	sid := ret.TableID
	var row rowResult
	err = sendGet(`row/contracts/`+sid, nil, &row)
	if err != nil {
		t.Error(err)
		return
	}
	code := row.Value[`value`]
	off := strings.IndexByte(code, '-')
	newCode := code[:off+1] + time.Now().Format(`2006.01.02`) + code[off+11:]
	form := url.Values{`Id`: {sid}, `Value`: {newCode}, `Conditions`: {row.Value[`conditions`]}, `WalletId`: {"01231234123412341230"}}
	if err := postTx(`EditContract`, &form); err != nil {
		t.Error(err)
		return
	}

	for _, item := range contracts {
		if !strings.HasSuffix(item.Name, `testUpd`) {
			continue
		}
		for _, par := range item.Params {
			form := url.Values{}
			for key, value := range par.Params {
				form[key] = []string{value}
			}
			if err := postTx(item.Name, &form); err != nil {
				t.Error(err)
				return
			}
			for key, value := range par.Results {
				if !wanted(key, value) {
					return
				}
			}
		}
	}
}

func TestActivateContracts(t *testing.T) {

	wanted := func(name, want string) bool {
		var ret getTestResult
		err := sendPost(`test/`+name, nil, &ret)
		if err != nil {
			t.Error(err)
			return false
		}
		if ret.Value != want {
			t.Error(fmt.Errorf(`%s != %s`, ret.Value, want))
			return false
		}
		return true
	}

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := `rnd` + crypto.RandSeq(6)
	form := url.Values{`Value`: {`contract ` + rnd + ` {
		    data {
				Par string
			}
			action { Test("active",  $Par)}}`}, `Conditions`: {`true`}}
	if err := postTx(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}
	var ret getContractResult
	err := sendGet(`contract/`+rnd, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if err := postTx(`ActivateContract`, &url.Values{`Id`: {ret.TableID}}); err != nil {
		t.Error(err)
		return
	}
	err = sendGet(`contract/`+rnd, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if !ret.Active {
		t.Error(fmt.Errorf(`Not activate ` + rnd))
	}
	var row rowResult
	err = sendGet(`row/contracts/`+ret.TableID, nil, &row)
	if err != nil {
		t.Error(err)
		return
	}
	if row.Value[`active`] != `1` {
		t.Error(fmt.Errorf(`row not activate ` + rnd))
	}

	if err := postTx(rnd, &url.Values{`Par`: {rnd}}); err != nil {
		t.Error(err)
		return
	}
	if !wanted(`active`, rnd) {
		return
	}
}

func TestDeactivateContracts(t *testing.T) {

	wanted := func(name, want string) bool {
		var ret getTestResult
		err := sendPost(`test/`+name, nil, &ret)
		if err != nil {
			t.Error(err)
			return false
		}
		if ret.Value != want {
			t.Error(fmt.Errorf(`%s != %s`, ret.Value, want))
			return false
		}
		return true
	}

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := `rnd` + crypto.RandSeq(6)
	form := url.Values{`Value`: {`contract ` + rnd + ` {
		    data {
				Par string
			}
			action { Test("active",  $Par)}}`}, `Conditions`: {`true`}}
	if err := postTx(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}
	var ret getContractResult
	err := sendGet(`contract/`+rnd, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if err := postTx(`ActivateContract`, &url.Values{`Id`: {ret.TableID}}); err != nil {
		t.Error(err)
		return
	}
	err = sendGet(`contract/`+rnd, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if !ret.Active {
		t.Error(fmt.Errorf(`Not activate ` + rnd))
	}
	var row rowResult
	err = sendGet(`row/contracts/`+ret.TableID, nil, &row)
	if err != nil {
		t.Error(err)
		return
	}
	if row.Value[`active`] != `1` {
		t.Error(fmt.Errorf(`row not activate ` + rnd))
	}

	if err := postTx(rnd, &url.Values{`Par`: {rnd}}); err != nil {
		t.Error(err)
		return
	}
	if !wanted(`active`, rnd) {
		return
	}

	if err := postTx(`DeactivateContract`, &url.Values{`Id`: {ret.TableID}}); err != nil {
		t.Error(err)
		return
	}
	err = sendGet(`contract/`+rnd, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if ret.Active {
		t.Error(fmt.Errorf(`Not deactivate ` + rnd))
	}
	var row2 rowResult
	err = sendGet(`row/contracts/`+ret.TableID, nil, &row2)
	if err != nil {
		t.Error(err)
		return
	}
	if row2.Value[`active`] != `0` {
		t.Error(fmt.Errorf(`row not deactivate ` + rnd))
	}
}

func TestContracts(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}

	var ret contractsResult
	err := sendGet(`contracts`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestSignature(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := `rnd` + crypto.RandSeq(6)
	form := url.Values{`Value`: {`contract ` + rnd + `Transfer {
		    data {
				Recipient int
				Amount    money
				Signature string "optional hidden"
			}
			action { 
				$result = "OK " + Str($Amount)
			}}
			
			contract ` + rnd + `Test {
				data {
					Recipient int "hidden"
					Amount  money
					Signature string "signature:` + rnd + `Transfer"
				}
				func action {
					` + rnd + `Transfer("Recipient,Amount,Signature",$Recipient,$Amount,$Signature )
					$result = "OOOPS " + Str($Amount)
				}
			  }
			`}, `Conditions`: {`true`}}
	if err := postTx(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Name`: {rnd + `Transfer`}, `Value`: {`{"title": "Would you like to sign",
		"params":[
			{"name": "Receipient", "text": "Wallet"},
			{"name": "Amount", "text": "Amount(EGS)"}
			]}`}, `Conditions`: {`true`}}
	if err := postTx(`NewSign`, &form); err != nil {
		t.Error(err)
		return
	}
	err := postTx(rnd+`Test`, &url.Values{`Amount`: {`12345`}, `Recipient`: {`98765`}})
	if err != nil {
		t.Error(err)
		return
	}
}

var (
	imp = `{
		"menus": [
			{
				"Name": "test_%s",
				"Conditions": "ContractAccess(\"@1EditMenu\")",
				"Value": "MenuItem(main, Default Ecosystem Menu)"
			}
		],
		"contracts": [
			{
				"Name": "testContract%[1]s",
				"Value": "contract testContract%[1]s {\n    data {}\n    conditions {}\n    action {\n        var res array\n        res = DBFind(\"pages\").Columns(\"name\").Where(\"id=?\", 1).Order(\"id\")\n        $result = res\n    }\n    }",
				"Conditions": "ContractConditions(` + "`MainCondition`" + `)"
			}
		],
		"pages": [
			{
				"Name": "test_%[1]s",
				"Conditions": "ContractAccess(\"@1EditPage\")",
				"Menu": "default_menu",
				"Value": "P(class, Default Ecosystem Page)\nImage().Style(width:100px;)"
			}
		],
		"blocks": [
			{
				"Name": "test_%[1]s",
				"Conditions": "true",
				"Value": "block content"
			},
			{
				"Name": "test_a%[1]s",
				"Conditions": "true",
				"Value": "block content"
			},
			{
				"Name": "test_b%[1]s",
				"Conditions": "true",
				"Value": "block content"
			}
		],
		"tables": [
			{
				"Name": "members%[1]s",
				"Columns": "[{\"name\":\"name\",\"type\":\"varchar\",\"conditions\":\"true\"},{\"name\":\"birthday\",\"type\":\"datetime\",\"conditions\":\"true\"},{\"name\":\"member_id\",\"type\":\"number\",\"conditions\":\"true\"},{\"name\":\"val\",\"type\":\"text\",\"conditions\":\"true\"},{\"name\":\"name_first\",\"type\":\"text\",\"conditions\":\"true\"},{\"name\":\"name_middle\",\"type\":\"text\",\"conditions\":\"true\"}]",
				"Permissions": "{\"insert\":\"true\",\"update\":\"true\",\"new_column\":\"true\"}"
			}
		],
		"parameters": [
			{
				"Name": "host%[1]s",
				"Value": "",
				"Conditions": "ContractConditions(` + "`MainCondition`" + `)"
			},
			{
				"Name": "host0%[1]s",
				"Value": "Русский текст",
				"Conditions": "ContractConditions(` + "`MainCondition`" + `)"
			}
		],
		"data": [
			{
				"Table": "members%[1]s",
				"Columns": ["name","val"],
				"Data": [
					["Bob","Richard mark"],
					["Mike Winter","Alan summer"]
				 ]
			}
		]
}`
)

func TestImport(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := crypto.RandSeq(4)
	form := url.Values{"Data": {fmt.Sprintf(imp, name)}}
	err := postTx(`@1Import`, &form)
	if err != nil {
		t.Error(err)
		return
	}

}

func TestEditContracts_ChangeWallet(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var cntlist contractsResult
	err := sendGet(`contracts`, nil, &cntlist)
	if err != nil {
		t.Error(err)
		return
	}
	var ret getContractResult
	err = sendGet(`contract/testUpd`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	sid := ret.TableID
	var row rowResult
	err = sendGet(`row/contracts/`+sid, nil, &row)
	if err != nil {
		t.Error(err)
		return
	}

	if err := postTx(`ActivateContract`, &url.Values{`Id`: {sid}}); err != nil {
		t.Error(err)
		return
	}

	code := row.Value[`value`]
	off := strings.IndexByte(code, '-')
	newCode := code[:off+1] + time.Now().Format(`2006.01.02`) + code[off+11:]
	form := url.Values{`Id`: {sid}, `Value`: {newCode}, `Conditions`: {row.Value[`conditions`]}, `WalletId`: {"1248-5499-7861-4204-5166"}}
	err = postTx(`EditContract`, &form)
	if err == nil {
		t.Error("Expected `Contract activated` error")
		return
	}

	if err := postTx(`DeactivateContract`, &url.Values{`Id`: {sid}}); err != nil {
		t.Error(err)
		return
	}

	if err := postTx(`EditContract`, &form); err != nil {
		t.Error(err)
		return
	}
}

func TestUpdateFunc(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}

	rnd := `rnd` + crypto.RandSeq(6)
	form := url.Values{`Value`: {`
		func MyTest(input string) string {
			return "X="+input
		}`}, `Conditions`: {`true`}}
	_, id, err := postTxResult(`NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}

	form = url.Values{`Value`: {`
		contract one` + rnd + ` {
			action {
				var ret map
				ret = DBFind("contracts").Columns("id,value").WhereId(10).Row()
				$result = ret["id"]
		}}
		contract row` + rnd + ` {
				action {
					var ret string
					ret = DBFind("contracts").Columns("id,value").WhereId(11).One("id")
					$result = ret
				}}
		
			`}, `Conditions`: {`true`}}
	err = postTx(`NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	_, msg, err := postTxResult(`one`+rnd, &url.Values{})
	if err != nil {
		t.Error(err)
		return
	}
	if msg != `10` {
		t.Error(`wrong one`)
		return
	}
	_, msg, err = postTxResult(`row`+rnd, &url.Values{})
	if err != nil {
		t.Error(err)
		return
	}
	if msg != `11` {
		t.Error(`wrong row`)
		return
	}

	form = url.Values{`Value`: {`
		contract ` + rnd + ` {
		    data {
				Par string
			}
			action {
				$result = MyTest($Par)
			}}
		`}, `Conditions`: {`true`}}
	_, idcnt, err := postTxResult(`NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	_, msg, err = postTxResult(rnd, &url.Values{`Par`: {`my param`}})
	if err != nil {
		t.Error(err)
		return
	}
	if msg != `X=my param` {
		t.Error(fmt.Errorf(`wrong result %s`, msg))
	}
	form = url.Values{`Id`: {id}, `Value`: {`
		func MyTest2(input string) string {
			return "Y="+input
		}`}, `Conditions`: {`true`}}
	err = postTx(`EditContract`, &form)
	if err.Error() != `{"type":"error","error":"Contracts or functions names cannot be changed"}` {
		t.Error(err)
		return
	}
	form = url.Values{`Id`: {id}, `Value`: {`
		func MyTest(input string) string {
			return "Y="+input
		}`}, `Conditions`: {`true`}}
	err = postTx(`EditContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	_, msg, err = postTxResult(rnd, &url.Values{`Par`: {`new param`}})
	if err != nil {
		t.Error(err)
		return
	}
	if msg != `Y=new param` {
		t.Errorf(`wrong result %s`, msg)
	}
	form = url.Values{`Id`: {idcnt}, `Value`: {`
		contract ` + rnd + ` {
		    data {
				Par string
			}
			action {
				$result = MyTest($Par) + MyTest("OK")
			}}
		`}, `Conditions`: {`true`}}
	_, idcnt, err = postTxResult(`EditContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	_, msg, err = postTxResult(rnd, &url.Values{`Par`: {`finish`}})
	if err != nil {
		t.Error(err)
		return
	}
	if msg != `Y=finishY=OK` {
		t.Errorf(`wrong result %s`, msg)
	}
}

func TestContractChain(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := `rnd` + crypto.RandSeq(4)

	form := url.Values{"Name": {rnd}, "Columns": {`[{"name":"value","type":"varchar", "index": "0", 
	  "conditions":"true"},
	{"name":"amount", "type":"number","index": "0", "conditions":"true"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`}}
	err := postTx(`NewTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Value`: {`contract sub` + rnd + ` {
		data {
			Id int
		}
		action {
			$row = DBFind("` + rnd + `").Columns("value").WhereId($Id)
			if Len($row) != 1 {
				error "sub contract getting error"
			}
			$record = $row[0]
			$new = $record["value"]
			DBUpdate("` + rnd + `", $Id, "value", $new+"="+$new )
		}
	}

	contract ` + rnd + ` {
		data {
			Initial string
		}
		action {
			$id = DBInsert("` + rnd + `", "value,amount", $Initial, "0")
			sub` + rnd + `("Id", $id)
			$row = DBFind("` + rnd + `").Columns("value").WhereId($id)
			if Len($row) != 1 {
				error "contract getting error"
			}
			$record = $row[0]
			$result = $record["value"]
		}
	}
		`}, `Conditions`: {`true`}}
	err = postTx(`NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	_, msg, err := postTxResult(rnd, &url.Values{`Initial`: {rnd}})
	if err != nil {
		t.Error(err)
		return
	}
	if msg != rnd+`=`+rnd {
		t.Error(fmt.Errorf(`wrong result %s`, msg))
	}
}
