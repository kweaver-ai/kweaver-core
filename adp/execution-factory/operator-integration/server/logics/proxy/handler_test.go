package proxy

// func TestXxx(t *testing.T) {
// 	pool := NewClientPool(10, 30*time.Second, 120*time.Second, 1*time.Second)
// 	forwarder := NewForwarder(pool)
// 	resp, err := forwarder.Forward(&interfaces.HTTPRequest{
// 		Method: "GET",
// 		URL:    "https://10.4.175.99/api/agent-operator-integration/v1/operator/info/{operator_id}",
// 		Headers: map[string]string{
// 			"Content-Type":  "application/json",
// 			"Authorization": "Bearer ory_at_o5v8zKf8fVhjTpq4C_MuOc0zAcnVEw1s8rnU5uFaX0E.4h_RzLXOSTR0Zw1rx9y73_J7a8HMExqseEPiaZvAKzo",
// 		},
// 		PathParams: map[string]string{
// 			"operator_id": "6bb72cc3-7666-4125-aedf-70a7c87e43da",
// 		},
// 		QueryParams: map[string]string{
// 			"version": "31f071d4-b951-4dad-b798-14efc9ac19a1",
// 		},
// 		ExecutionMode: "sync",
// 		Timeout:       30 * time.Second,
// 	}, "123")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	t.Log(resp)
// }
