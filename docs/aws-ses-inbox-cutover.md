# AWS SES + Admin Inbox Cutover

## 1. AWS SES domain identity
- Region: `ap-southeast-3`
- Domain identity: `ppob.id`
- Enable `Easy DKIM`
- Custom MAIL FROM domain: `bounce.ppob.id`
- Configuration sets:
  - `ppob-transactional`
  - `ppob-operations`

## 2. Cloudflare DNS
Set all email-related records to `DNS only`.

Add the SES records:
- DKIM CNAME records from SES verification
- MAIL FROM MX for `bounce.ppob.id`
- MAIL FROM TXT for `bounce.ppob.id`
- SPF TXT for `ppob.id`
- DMARC TXT for `_dmarc.ppob.id`
- SES inbound MX records for `ppob.id`

## 3. AWS SES production access
- Request production access for SES in `ap-southeast-3`
- Verify `ppob.id` status is `Verified`
- Verify MAIL FROM status is `Success`

## 4. S3 + SNS for inbound email
- Create S3 bucket for raw inbound email
- Create SNS topic for inbound notifications
- Create SNS topic for outbound delivery events
- Create SES receipt rule:
  - store inbound raw email to S3
  - publish receipt metadata to inbound SNS
  - apply to catch-all `@ppob.id`

## 5. Backend environment
Set these env vars before backend deploy:

```env
EMAIL_PROVIDER=ses
EMAIL_FROM_DEFAULT=noreply@ppob.id
EMAIL_FROM_NAME=PPOB.ID
EMAIL_REPLY_TO=cs@ppob.id
EMAIL_MAILBOX_DOMAIN=ppob.id

SES_REGION=ap-southeast-3
SES_ACCESS_KEY_ID=...
SES_SECRET_ACCESS_KEY=...
SES_CONFIGURATION_SET_TRANSACTIONAL=ppob-transactional
SES_CONFIGURATION_SET_OPERATIONS=ppob-operations
SES_MAIL_FROM_DOMAIN=bounce.ppob.id
SES_INBOUND_BUCKET=...
SES_INBOUND_TOPIC_ARN=...
SES_DELIVERY_TOPIC_ARN=...
```

## 6. Backend endpoints
These endpoints must be reachable from AWS SNS:
- `POST /v1/admin/email/inbound/sns`
- `POST /v1/admin/email/delivery-events/sns`

## 7. Deploy order
1. Deploy backend migration + code
2. Verify admin inbox routes respond
3. Deploy admin panel
4. Send transactional test from `noreply@ppob.id`
5. Send inbound test to `cs@ppob.id`
6. Confirm delivery/bounce/complaint events reach `email_dispatch_logs`

## 8. Validation checklist
- Admin invite email sent via SES
- Admin forgot password email sent via SES
- `mailbox_reply` log appears in `email_dispatch_logs`
- Inbound email creates thread in admin inbox
- Personal mailbox only visible to owner + executive roles
- `unmapped@ppob.id` only visible to executive roles

## 9. Rollback
- Set `EMAIL_PROVIDER=brevo`
- Re-deploy backend
- Keep SES DNS records in place if needed for retry window
