package whatsapp

import "testing"

func TestVerifySignature(t *testing.T) {
	payload := []byte(`{"object":"whatsapp_business_account"}`)
	signature := "sha256=9f7bc7d657c97d1250a31b35f3f4ecdee2ce26677e377b89309d18f2f0f2f6aa"

	if !VerifySignature(payload, signature, "super-secret") {
		t.Fatal("expected signature to be valid")
	}

	if VerifySignature(payload, signature, "wrong-secret") {
		t.Fatal("expected signature to be invalid with wrong secret")
	}
}

func TestParseWebhook(t *testing.T) {
	payload := []byte(`{
		"object":"whatsapp_business_account",
		"entry":[
			{
				"id":"123",
				"changes":[
					{
						"field":"messages",
						"value":{
							"metadata":{
								"display_phone_number":"+62 813-8181-650",
								"phone_number_id":"965594676640862"
							},
							"messages":[
								{
									"from":"628138181630",
									"id":"wamid-message",
									"timestamp":"1710000000",
									"type":"text",
									"text":{"body":"tes"}
								}
							],
							"statuses":[
								{
									"id":"wamid-status",
									"status":"delivered",
									"timestamp":"1710000001",
									"recipient_id":"628138181630"
								}
							]
						}
					}
				]
			}
		]
	}`)

	webhook, err := ParseWebhook(payload)
	if err != nil {
		t.Fatalf("expected webhook to parse, got error: %v", err)
	}

	if webhook.Object != "whatsapp_business_account" {
		t.Fatalf("unexpected object: %s", webhook.Object)
	}

	if got := len(webhook.Entry); got != 1 {
		t.Fatalf("expected 1 entry, got %d", got)
	}

	change := webhook.Entry[0].Changes[0]
	if change.Field != "messages" {
		t.Fatalf("unexpected field: %s", change.Field)
	}

	if got := len(change.Value.Messages); got != 1 {
		t.Fatalf("expected 1 inbound message, got %d", got)
	}

	if got := len(change.Value.Statuses); got != 1 {
		t.Fatalf("expected 1 status, got %d", got)
	}
}
