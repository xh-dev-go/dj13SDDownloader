package main

import (
	"testing"
)

func Test_main(t *testing.T) {
	v := `
title Stocking Flow - 2.1 System Update Stock Request
begin User as user,StockRequestService as sr, BatchService as bs, OrderSystem as os, Scheduler as scheduler
User is a person

scheduler -->> +bs: execute stock request process
bs -> +os: "execute stock request process\napi: **_/api/orders/{since}_**"
if "[has stock request]"
  -os --> bs: return any stock request
  activate sr
  repeat addStockRequest:
    bs -> sr: add stock request
  end
deactivate sr
end
deactivate bs


+user -> +bs: **_getPendingRequest()_**
-bs --> user: list of pending request
+user -> +bs: **_confirmPendingRequest()_**
-bs --> -user: result`
	r := `title%20Stocking%20Flow%20-%202.1%20System%20Update%20Stock%20Request/begin%20User%20as%20user%2CStockRequestService%20as%20sr%2C%20BatchService%20as%20bs%2C%20OrderSystem%20as%20os%2C%20Scheduler%20as%20scheduler/User%20is%20a%20person/scheduler%20--%3E%3E%20%2Bbs%3A%20execute%20stock%20request%20process/bs%20-%3E%20%2Bos%3A%20%22execute%20stock%20request%20process%5Cnapi%3A%20**_%2Fapi%2Forders%2F%7Bsince%7D_**%22/if%20%22%5Bhas%20stock%20request%5D%22/%20%20-os%20--%3E%20bs%3A%20return%20any%20stock%20request/%20%20activate%20sr/%20%20repeat%20addStockRequest%3A/%20%20%20%20bs%20-%3E%20sr%3A%20add%20stock%20request/%20%20end/deactivate%20sr/end/deactivate%20bs/%2Buser%20-%3E%20%2Bbs%3A%20**_getPendingRequest()_**/-bs%20--%3E%20user%3A%20list%20of%20pending%20request/%2Buser%20-%3E%20%2Bbs%3A%20**_confirmPendingRequest()_**/-bs%20--%3E%20-user%3A%20result`
	v = ProcessScript(v)
	if v != r {
		t.Fail()
	}

}
