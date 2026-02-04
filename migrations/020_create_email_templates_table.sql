-- Migration: 020_create_email_templates_table
-- Description: Create email_templates table for dynamic email content management
-- Created: 2025-01-31
-- Note: Email hanya untuk auth & security (verifikasi email, login baru, PIN changed, email changed)
--       dan marketing. BUKAN untuk OTP atau notifikasi transaksi.

CREATE TABLE IF NOT EXISTS email_templates (
    id VARCHAR(36) PRIMARY KEY,                           -- UUID: emt_xxx
    code VARCHAR(50) UNIQUE NOT NULL,                     -- Template code: EMAIL_VERIFICATION, SECURITY_NEW_LOGIN, etc.
    name VARCHAR(100) NOT NULL,                           -- Display name
    description TEXT,                                     -- Template description
    category VARCHAR(30) NOT NULL,                        -- AUTH, SECURITY, MARKETING

    -- Email Content
    subject VARCHAR(255) NOT NULL,                        -- Email subject (supports {{params}})
    html_content TEXT NOT NULL,                           -- HTML template content
    text_content TEXT,                                    -- Plain text fallback

    -- Brevo Integration
    brevo_template_id INT,                                -- Brevo template ID (if using Brevo templates)

    -- Template Variables
    variables JSONB DEFAULT '[]',                         -- List of variables: [{"name": "USER_NAME", "type": "string", "required": true}]

    -- Settings
    is_active BOOLEAN DEFAULT TRUE,
    priority INT DEFAULT 0,                               -- Higher = more important

    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(36),                               -- Admin user ID
    updated_by VARCHAR(36)
);

-- Indexes
CREATE INDEX idx_email_templates_code ON email_templates(code);
CREATE INDEX idx_email_templates_category ON email_templates(category);
CREATE INDEX idx_email_templates_is_active ON email_templates(is_active);

-- Insert default templates (AUTH & SECURITY only)
INSERT INTO email_templates (id, code, name, description, category, subject, html_content, text_content, brevo_template_id, variables, is_active, priority) VALUES

-- 1. Email Verification
('emt_email_verification', 'EMAIL_VERIFICATION', 'Verifikasi Email', 'Template untuk mengirim link verifikasi email', 'AUTH',
'Verifikasi Email Anda',
'<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verifikasi Email Anda</title>
    <style>
        body { margin: 0; padding: 0; -webkit-text-size-adjust: 100%; -ms-text-size-adjust: 100%; background-color: #F0F4F8; }
        table { border-spacing: 0; border-collapse: collapse; }
        img { border: 0; line-height: 100%; outline: none; text-decoration: none; display: block; }

        body, td { font-family: ''Helvetica Neue'', Helvetica, Arial, sans-serif; color: #333333; }

        .btn-primary {
            background-color: #007BFF;
            color: #ffffff !important;
            padding: 12px 24px;
            border-radius: 6px;
            text-decoration: none;
            font-weight: bold;
            display: inline-block;
            mso-padding-alt: 0;
            text-align: center;
        }

        .footer-link:hover { opacity: 0.8 !important; }

        @media only screen and (max-width: 600px) {
            .container { width: 100% !important; padding: 0 !important; }
            .content-padding { padding: 20px !important; }
            .store-btn { width: 130px !important; height: auto !important; }
            .separator-padding { padding: 0 20px !important; }
        }
    </style>
</head>
<body style="background-color: #F0F4F8; padding: 40px 0;">
<table width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #F0F4F8;">
    <tr>
        <td align="center">
            <table class="container" width="600" border="0" cellspacing="0" cellpadding="0" style="background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 15px rgba(0,0,0,0.05);">
                <tr><td height="6" style="background-color: #007BFF;"></td></tr>
                <tr>
                    <td align="center" style="padding: 40px 0 20px 0;">
                        <img src="https://cdn.ppob.id/server/logo.png" alt="Logo PPOB.id" width="150" style="display: block;">
                    </td>
                </tr>
                <tr>
                    <td class="separator-padding" style="padding: 0 50px;">
                        <div style="border-top: 1px solid #E9ECEF; height: 1px; line-height: 1px; width: 100%;"></div>
                    </td>
                </tr>
                <tr>
                    <td class="content-padding" style="padding: 20px 50px;">
                        <h1 style="color: #004085; font-size: 24px; margin-bottom: 20px; text-align: center;">
                            Verifikasi Email Anda
                        </h1>
                        <p style="font-size: 16px; line-height: 1.6; color: #555555; text-align: center; margin-bottom: 30px;">
                            Halo, <strong>{{USER_NAME}}</strong><br><br>
                            Untuk meningkatkan keamanan akun Anda, silakan verifikasi alamat email Anda dengan menekan tombol di bawah ini:
                        </p>
                        <table width="100%" border="0" cellspacing="0" cellpadding="0">
                            <tr>
                                <td align="center" style="padding-bottom: 30px;">
                                    <a href="{{VERIFICATION_LINK}}" class="btn-primary">Verifikasi Email Sekarang</a>
                                </td>
                            </tr>
                        </table>
                        <p style="font-size: 13px; color: #999999; text-align: center; margin-top: 0; margin-bottom: 10px;">
                            Jika Anda tidak merasa melakukan permintaan ini, silakan abaikan email ini.
                        </p>
                    </td>
                </tr>
                <tr>
                    <td class="content-padding" style="padding: 0 50px 40px 50px;">
                        <table width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #EBF5FF; border-radius: 8px; border: 1px solid #D6E9FF;">
                            <tr>
                                <td align="center" style="padding: 20px;">
                                    <p style="margin: 0; font-size: 14px; color: #004085; font-weight: bold;">Butuh Bantuan?</p>
                                    <p style="margin: 10px 0 0 0; font-size: 14px; color: #333333;">
                                        📧 <a href="mailto:cs@ppob.id" style="color: #007BFF; text-decoration: none;">cs@ppob.id</a> &nbsp;|&nbsp;
                                        📱 <a href="https://wa.me/628138181640" style="color: #007BFF; text-decoration: none;">+62 813-8181-640</a>
                                    </p>
                                </td>
                            </tr>
                        </table>
                    </td>
                </tr>
                <tr>
                    <td class="separator-padding" style="padding: 0 50px;">
                        <div style="border-top: 1px solid #E9ECEF; height: 1px; line-height: 1px; width: 100%;"></div>
                    </td>
                </tr>
                <tr>
                    <td style="padding: 35px 40px; color: #333333; text-align: center;">
                        <p style="margin: 0; font-weight: bold; font-size: 15px; letter-spacing: 0.5px; color: #004085;">PT Gerbang Transaksi Digital</p>
                        <p style="margin: 8px 0 25px 0; font-size: 12px; color: #666666; line-height: 1.5;">
                            Wisma KEIAI 14th Floor Unit 1410<br>
                            Jln. Jenderal Sudirman Kav.3, Karet Tengsin, Tanah Abang Jakarta Pusat, Jakarta, 10220
                        </p>
                        <p style="margin: 0 0 30px 0;">
                            <a href="https://facebook.com/ppob.id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-facebook.svg" alt="Facebook" width="28" height="28" style="display: block; border: 0;">
                            </a>
                            <a href="https://instagram.com/ppob.id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-instagram.svg" alt="Instagram" width="28" height="28" style="display: block; border: 0;">
                            </a>
                            <a href="https://twitter.com/ppob_id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-x.svg" alt="Twitter X" width="28" height="28" style="display: block; border: 0;">
                            </a>
                            <a href="https://tiktok.com/@ppob.id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-tiktok.svg" alt="TikTok" width="28" height="28" style="display: block; border: 0;">
                            </a>
                        </p>
                        <table width="100%" border="0" cellspacing="0" cellpadding="0">
                            <tr>
                                <td align="center">
                                    <p style="font-size: 13px; font-weight: bold; color: #333333; margin: 0 0 15px 0; letter-spacing: 0.5px;">
                                        Unduh ppob.id melalui
                                    </p>
                                </td>
                            </tr>
                            <tr>
                                <td align="center" style="padding-bottom: 25px;">
                                    <table border="0" cellspacing="0" cellpadding="0" style="margin: 0 auto;">
                                        <tr>
                                            <td valign="middle" style="padding-right: 15px;">
                                                <img src="https://cdn.ppob.id/server/icon.png" alt="Icon" width="42" height="42" style="display: block; border-radius: 8px;">
                                            </td>
                                            <td valign="middle" style="padding-right: 8px;">
                                                <a href="#">
                                                    <img src="https://cdn.ppob.id/server/google-play.png" alt="Google Play" height="35" style="display: block; border:0;" class="store-btn">
                                                </a>
                                            </td>
                                            <td valign="middle">
                                                <a href="#">
                                                    <img src="https://cdn.ppob.id/server/app-store.png" alt="App Store" height="35" style="display: block; border:0;" class="store-btn">
                                                </a>
                                            </td>
                                        </tr>
                                    </table>
                                </td>
                            </tr>
                        </table>
                        <p style="margin: 0; font-size: 11px; color: #999999;">
                            &copy; 2026 PT Gerbang Transaksi Digital. All rights reserved.
                        </p>
                    </td>
                </tr>
            </table>
            <table width="100%" border="0" cellspacing="0" cellpadding="0">
                <tr><td height="40"></td></tr>
            </table>
        </td>
    </tr>
</table>
</body>
</html>',
'Halo, {{USER_NAME}}

Untuk meningkatkan keamanan akun Anda, silakan verifikasi alamat email Anda dengan menekan tombol di bawah ini:
{{VERIFICATION_LINK}}

Jika Anda tidak merasa melakukan permintaan ini, silakan abaikan email ini.

ppob.id',
1,
'[{"name": "USER_NAME", "type": "string", "required": true}, {"name": "VERIFICATION_LINK", "type": "string", "required": true}, {"name": "EXPIRES_IN", "type": "string", "required": true, "default": "24 jam"}]',
TRUE, 100),

-- 2. Security Alert - New Login
('emt_security_new_login', 'SECURITY_NEW_LOGIN', 'Peringatan Login Baru', 'Notifikasi ketika ada login dari perangkat baru', 'SECURITY',
'Peringatan Keamanan: Login Baru Terdeteksi',
'<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login Baru Terdeteksi</title>
    <style>
        body { margin: 0; padding: 0; -webkit-text-size-adjust: 100%; -ms-text-size-adjust: 100%; background-color: #F0F4F8; }
        table { border-spacing: 0; border-collapse: collapse; }
        img { border: 0; line-height: 100%; outline: none; text-decoration: none; display: block; }

        body, td { font-family: ''Helvetica Neue'', Helvetica, Arial, sans-serif; color: #333333; }

        .btn-primary {
            background-color: #007BFF;
            color: #ffffff !important;
            padding: 12px 24px;
            border-radius: 6px;
            text-decoration: none;
            font-weight: bold;
            display: inline-block;
            mso-padding-alt: 0;
            text-align: center;
        }

        .footer-link:hover { opacity: 0.8 !important; }

        @media only screen and (max-width: 600px) {
            .container { width: 100% !important; padding: 0 !important; }
            .content-padding { padding: 20px !important; }
            .store-btn { width: 130px !important; height: auto !important; }
            .separator-padding { padding: 0 20px !important; }
        }
    </style>
</head>
<body style="background-color: #F0F4F8; padding: 40px 0;">
<table width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #F0F4F8;">
    <tr>
        <td align="center">
            <table class="container" width="600" border="0" cellspacing="0" cellpadding="0" style="background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 15px rgba(0,0,0,0.05);">
                <tr><td height="6" style="background-color: #007BFF;"></td></tr>
                <tr>
                    <td align="center" style="padding: 40px 0 20px 0;">
                        <img src="https://cdn.ppob.id/server/logo.png" alt="Logo PPOB.id" width="150" style="display: block;">
                    </td>
                </tr>
                <tr>
                    <td class="separator-padding" style="padding: 0 50px;">
                        <div style="border-top: 1px solid #E9ECEF; height: 1px; line-height: 1px; width: 100%;"></div>
                    </td>
                </tr>
                <tr>
                    <td class="content-padding" style="padding: 20px 50px;">
                        <h1 style="color: #004085; font-size: 24px; margin-bottom: 20px; text-align: center;">
                            Login Baru Terdeteksi
                        </h1>
                        <p style="font-size: 16px; line-height: 1.6; color: #555555; text-align: center; margin-bottom: 25px;">
                            Halo, <strong>{{USER_NAME}}</strong><br><br>
                            Kami mendeteksi perangkat baru masuk ke akun ppob.id Anda. Jika ini adalah Anda, silakan abaikan pesan ini.
                        </p>
                        <table width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #F8F9FA; border-radius: 8px; border: 1px solid #E9ECEF; margin-bottom: 25px;">
                            <tr>
                                <td style="padding: 15px 20px;">
                                    <table width="100%" border="0" cellspacing="0" cellpadding="0">
                                        <tr>
                                            <td style="font-size: 13px; color: #999999; padding-bottom: 5px;">Perangkat</td>
                                            <td align="right" style="font-size: 13px; color: #333333; font-weight: bold;">{{DEVICE_NAME}}</td>
                                        </tr>
                                        <tr>
                                            <td style="font-size: 13px; color: #999999; padding-bottom: 5px;">Lokasi</td>
                                            <td align="right" style="font-size: 13px; color: #333333; font-weight: bold;">{{LOCATION}}</td>
                                        </tr>
                                        <tr>
                                            <td style="font-size: 13px; color: #999999; padding-bottom: 5px;">Alamat IP</td>
                                            <td align="right" style="font-size: 13px; color: #333333; font-weight: bold;">{{IP_ADDRESS}}</td>
                                        </tr>
                                        <tr>
                                            <td style="font-size: 13px; color: #999999; padding-bottom: 5px;">Waktu</td>
                                            <td align="right" style="font-size: 13px; color: #333333; font-weight: bold;">{{LOGIN_TIME}}</td>
                                        </tr>
                                    </table>
                                </td>
                            </tr>
                        </table>
                        <p style="font-size: 14px; line-height: 1.5; color: #333333; text-align: center; margin-bottom: 20px;">
                            Jika ini <strong>bukan Anda</strong>, segera amankan akun Anda.
                        </p>
                    </td>
                </tr>
                <tr>
                    <td class="content-padding" style="padding: 0 50px 40px 50px;">
                        <table width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #EBF5FF; border-radius: 8px; border: 1px solid #D6E9FF;">
                            <tr>
                                <td align="center" style="padding: 20px;">
                                    <p style="margin: 0; font-size: 14px; color: #004085; font-weight: bold;">Butuh Bantuan?</p>
                                    <p style="margin: 10px 0 0 0; font-size: 14px; color: #333333;">
                                        📧 <a href="mailto:cs@ppob.id" style="color: #007BFF; text-decoration: none;">cs@ppob.id</a> &nbsp;|&nbsp;
                                        📱 <a href="https://wa.me/628138181640" style="color: #007BFF; text-decoration: none;">+62 813-8181-640</a>
                                    </p>
                                </td>
                            </tr>
                        </table>
                    </td>
                </tr>
                <tr>
                    <td class="separator-padding" style="padding: 0 50px;">
                        <div style="border-top: 1px solid #E9ECEF; height: 1px; line-height: 1px; width: 100%;"></div>
                    </td>
                </tr>
                <tr>
                    <td style="padding: 35px 40px; color: #333333; text-align: center;">
                        <p style="margin: 0; font-weight: bold; font-size: 15px; letter-spacing: 0.5px; color: #004085;">PT Gerbang Transaksi Digital</p>
                        <p style="margin: 8px 0 25px 0; font-size: 12px; color: #666666; line-height: 1.5;">
                            Wisma KEIAI 14th Floor Unit 1410<br>
                            Jln. Jenderal Sudirman Kav.3, Karet Tengsin, Tanah Abang Jakarta Pusat, Jakarta, 10220
                        </p>
                        <p style="margin: 0 0 30px 0;">
                            <a href="https://facebook.com/ppob.id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-facebook.svg" alt="Facebook" width="28" height="28" style="display: block; border: 0;">
                            </a>
                            <a href="https://instagram.com/ppob.id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-instagram.svg" alt="Instagram" width="28" height="28" style="display: block; border: 0;">
                            </a>
                            <a href="https://twitter.com/ppob_id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-x.svg" alt="Twitter X" width="28" height="28" style="display: block; border: 0;">
                            </a>
                            <a href="https://tiktok.com/@ppob.id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-tiktok.svg" alt="TikTok" width="28" height="28" style="display: block; border: 0;">
                            </a>
                        </p>
                        <table width="100%" border="0" cellspacing="0" cellpadding="0">
                            <tr>
                                <td align="center">
                                    <p style="font-size: 13px; font-weight: bold; color: #333333; margin: 0 0 15px 0; letter-spacing: 0.5px;">
                                        Unduh ppob.id melalui
                                    </p>
                                </td>
                            </tr>
                            <tr>
                                <td align="center" style="padding-bottom: 25px;">
                                    <table border="0" cellspacing="0" cellpadding="0" style="margin: 0 auto;">
                                        <tr>
                                            <td valign="middle" style="padding-right: 15px;">
                                                <img src="https://cdn.ppob.id/server/icon.png" alt="Icon" width="42" height="42" style="display: block; border-radius: 8px;">
                                            </td>
                                            <td valign="middle" style="padding-right: 8px;">
                                                <a href="#">
                                                    <img src="https://cdn.ppob.id/server/google-play.png" alt="Google Play" height="35" style="display: block; border:0;" class="store-btn">
                                                </a>
                                            </td>
                                            <td valign="middle">
                                                <a href="#">
                                                    <img src="https://cdn.ppob.id/server/app-store.png" alt="App Store" height="35" style="display: block; border:0;" class="store-btn">
                                                </a>
                                            </td>
                                        </tr>
                                    </table>
                                </td>
                            </tr>
                        </table>
                        <p style="margin: 0; font-size: 11px; color: #999999;">
                            &copy; 2026 PT Gerbang Transaksi Digital. All rights reserved.
                        </p>
                    </td>
                </tr>
            </table>
            <table width="100%" border="0" cellspacing="0" cellpadding="0">
                <tr><td height="40"></td></tr>
            </table>
        </td>
    </tr>
</table>
</body>
</html>',
'Halo, {{USER_NAME}}

Kami mendeteksi perangkat baru masuk ke akun ppob.id Anda. Jika ini adalah Anda, silakan abaikan pesan ini.

Akun ppob.id Anda login dari perangkat baru:
- Perangkat: {{DEVICE_NAME}}
- Lokasi: {{LOCATION}}
- IP Address: {{IP_ADDRESS}}
- Waktu: {{LOGIN_TIME}}

Jika ini bukan Anda, segera amankan akun Anda.

ppob.id security',
2,
'[{"name": "USER_NAME", "type": "string", "required": true}, {"name": "DEVICE_NAME", "type": "string", "required": true}, {"name": "LOCATION", "type": "string", "required": true}, {"name": "IP_ADDRESS", "type": "string", "required": true}, {"name": "LOGIN_TIME", "type": "string", "required": true}]',
TRUE, 90),

-- 3. Security Alert - PIN Changed
('emt_security_pin_changed', 'SECURITY_PIN_CHANGED', 'PIN Diubah', 'Notifikasi ketika PIN berhasil diubah', 'SECURITY',
'Perubahan PIN Berhasil',
'<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Perubahan PIN Berhasil</title>
    <style>
        body { margin: 0; padding: 0; -webkit-text-size-adjust: 100%; -ms-text-size-adjust: 100%; background-color: #F0F4F8; }
        table { border-spacing: 0; border-collapse: collapse; }
        img { border: 0; line-height: 100%; outline: none; text-decoration: none; display: block; }

        body, td { font-family: ''Helvetica Neue'', Helvetica, Arial, sans-serif; color: #333333; }

        .btn-primary {
            background-color: #007BFF;
            color: #ffffff !important;
            padding: 12px 24px;
            border-radius: 6px;
            text-decoration: none;
            font-weight: bold;
            display: inline-block;
            mso-padding-alt: 0;
            text-align: center;
        }

        .footer-link:hover { opacity: 0.8 !important; }

        @media only screen and (max-width: 600px) {
            .container { width: 100% !important; padding: 0 !important; }
            .content-padding { padding: 20px !important; }
            .store-btn { width: 130px !important; height: auto !important; }
            .separator-padding { padding: 0 20px !important; }
        }
    </style>
</head>
<body style="background-color: #F0F4F8; padding: 40px 0;">
<table width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #F0F4F8;">
    <tr>
        <td align="center">
            <table class="container" width="600" border="0" cellspacing="0" cellpadding="0" style="background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 15px rgba(0,0,0,0.05);">
                <tr><td height="6" style="background-color: #007BFF;"></td></tr>
                <tr>
                    <td align="center" style="padding: 40px 0 20px 0;">
                        <img src="https://cdn.ppob.id/server/logo.png" alt="Logo PPOB.id" width="150" style="display: block;">
                    </td>
                </tr>
                <tr>
                    <td class="separator-padding" style="padding: 0 50px;">
                        <div style="border-top: 1px solid #E9ECEF; height: 1px; line-height: 1px; width: 100%;"></div>
                    </td>
                </tr>
                <tr>
                    <td class="content-padding" style="padding: 20px 50px;">
                        <h1 style="color: #004085; font-size: 24px; margin-bottom: 20px; text-align: center;">
                            Perubahan PIN Berhasil
                        </h1>
                        <p style="font-size: 16px; line-height: 1.6; color: #555555; text-align: center; margin-bottom: 25px;">
                            Halo, <strong>{{USER_NAME}}</strong><br><br>
                            PIN akun ppob.id Anda berhasil diubah.
                        </p>
                        <table width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #F8F9FA; border-radius: 8px; border: 1px solid #E9ECEF; margin-bottom: 25px;">
                            <tr>
                                <td style="padding: 15px 20px;">
                                    <table width="100%" border="0" cellspacing="0" cellpadding="0">
                                        <tr>
                                            <td style="font-size: 13px; color: #999999; padding-bottom: 5px;">Perangkat</td>
                                            <td align="right" style="font-size: 13px; color: #333333; font-weight: bold;">{{DEVICE_NAME}}</td>
                                        </tr>
                                        <tr>
                                            <td style="font-size: 13px; color: #999999; padding-bottom: 5px;">Waktu Perubahan</td>
                                            <td align="right" style="font-size: 13px; color: #333333; font-weight: bold;">{{CHANGE_TIME}}</td>
                                        </tr>
                                    </table>
                                </td>
                            </tr>
                        </table>
                        <p style="font-size: 14px; line-height: 1.5; color: #333333; text-align: center; margin-bottom: 20px;">
                            Jika ini <strong>bukan Anda</strong>, segera amankan akun Anda.
                        </p>
                    </td>
                </tr>
                <tr>
                    <td class="content-padding" style="padding: 0 50px 40px 50px;">
                        <table width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #EBF5FF; border-radius: 8px; border: 1px solid #D6E9FF;">
                            <tr>
                                <td align="center" style="padding: 20px;">
                                    <p style="margin: 0; font-size: 14px; color: #004085; font-weight: bold;">Butuh Bantuan?</p>
                                    <p style="margin: 10px 0 0 0; font-size: 14px; color: #333333;">
                                        📧 <a href="mailto:cs@ppob.id" style="color: #007BFF; text-decoration: none;">cs@ppob.id</a> &nbsp;|&nbsp;
                                        📱 <a href="https://wa.me/628138181640" style="color: #007BFF; text-decoration: none;">+62 813-8181-640</a>
                                    </p>
                                </td>
                            </tr>
                        </table>
                    </td>
                </tr>
                <tr>
                    <td class="separator-padding" style="padding: 0 50px;">
                        <div style="border-top: 1px solid #E9ECEF; height: 1px; line-height: 1px; width: 100%;"></div>
                    </td>
                </tr>
                <tr>
                    <td style="padding: 35px 40px; color: #333333; text-align: center;">
                        <p style="margin: 0; font-weight: bold; font-size: 15px; letter-spacing: 0.5px; color: #004085;">PT Gerbang Transaksi Digital</p>
                        <p style="margin: 8px 0 25px 0; font-size: 12px; color: #666666; line-height: 1.5;">
                            Wisma KEIAI 14th Floor Unit 1410<br>
                            Jln. Jenderal Sudirman Kav.3, Karet Tengsin, Tanah Abang Jakarta Pusat, Jakarta, 10220
                        </p>
                        <p style="margin: 0 0 30px 0;">
                            <a href="https://facebook.com/ppob.id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-facebook.svg" alt="Facebook" width="28" height="28" style="display: block; border: 0;">
                            </a>
                            <a href="https://instagram.com/ppob.id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-instagram.svg" alt="Instagram" width="28" height="28" style="display: block; border: 0;">
                            </a>
                            <a href="https://twitter.com/ppob_id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-x.svg" alt="Twitter X" width="28" height="28" style="display: block; border: 0;">
                            </a>
                            <a href="https://tiktok.com/@ppob.id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-tiktok.svg" alt="TikTok" width="28" height="28" style="display: block; border: 0;">
                            </a>
                        </p>
                        <table width="100%" border="0" cellspacing="0" cellpadding="0">
                            <tr>
                                <td align="center">
                                    <p style="font-size: 13px; font-weight: bold; color: #333333; margin: 0 0 15px 0; letter-spacing: 0.5px;">
                                        Unduh ppob.id melalui
                                    </p>
                                </td>
                            </tr>
                            <tr>
                                <td align="center" style="padding-bottom: 25px;">
                                    <table border="0" cellspacing="0" cellpadding="0" style="margin: 0 auto;">
                                        <tr>
                                            <td valign="middle" style="padding-right: 15px;">
                                                <img src="https://cdn.ppob.id/server/icon.png" alt="Icon" width="42" height="42" style="display: block; border-radius: 8px;">
                                            </td>
                                            <td valign="middle" style="padding-right: 8px;">
                                                <a href="#">
                                                    <img src="https://cdn.ppob.id/server/google-play.png" alt="Google Play" height="35" style="display: block; border:0;" class="store-btn">
                                                </a>
                                            </td>
                                            <td valign="middle">
                                                <a href="#">
                                                    <img src="https://cdn.ppob.id/server/app-store.png" alt="App Store" height="35" style="display: block; border:0;" class="store-btn">
                                                </a>
                                            </td>
                                        </tr>
                                    </table>
                                </td>
                            </tr>
                        </table>
                        <p style="margin: 0; font-size: 11px; color: #999999;">
                            &copy; 2026 PT Gerbang Transaksi Digital. All rights reserved.
                        </p>
                    </td>
                </tr>
            </table>
            <table width="100%" border="0" cellspacing="0" cellpadding="0">
                <tr><td height="40"></td></tr>
            </table>
        </td>
    </tr>
</table>
</body>
</html>',
'Halo, {{USER_NAME}}

PIN akun ppob.id Anda berhasil diubah.
' ||
'Perangkat: {{DEVICE_NAME}}
Waktu Perubahan: {{CHANGE_TIME}}

Jika ini bukan Anda, segera amankan akun Anda.

ppob.id security',
3,
'[{"name": "USER_NAME", "type": "string", "required": true}, {"name": "CHANGE_TIME", "type": "string", "required": true}, {"name": "DEVICE_NAME", "type": "string", "required": true}]',
TRUE, 90),

-- 4. Security Alert - Email Changed
('emt_security_email_changed', 'SECURITY_EMAIL_CHANGED', 'Email Diubah', 'Notifikasi ketika email diubah', 'SECURITY',
'Alamat Email Berhasil Diubah',
'<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Alamat Email Berhasil Diubah</title>
    <style>
        body { margin: 0; padding: 0; -webkit-text-size-adjust: 100%; -ms-text-size-adjust: 100%; background-color: #F0F4F8; }
        table { border-spacing: 0; border-collapse: collapse; }
        img { border: 0; line-height: 100%; outline: none; text-decoration: none; display: block; }

        body, td { font-family: ''Helvetica Neue'', Helvetica, Arial, sans-serif; color: #333333; }

        .btn-primary {
            background-color: #007BFF;
            color: #ffffff !important;
            padding: 12px 24px;
            border-radius: 6px;
            text-decoration: none;
            font-weight: bold;
            display: inline-block;
            mso-padding-alt: 0;
            text-align: center;
        }

        .footer-link:hover { opacity: 0.8 !important; }

        @media only screen and (max-width: 600px) {
            .container { width: 100% !important; padding: 0 !important; }
            .content-padding { padding: 20px !important; }
            .store-btn { width: 130px !important; height: auto !important; }
            .separator-padding { padding: 0 20px !important; }
        }
    </style>
</head>
<body style="background-color: #F0F4F8; padding: 40px 0;">
<table width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #F0F4F8;">
    <tr>
        <td align="center">
            <table class="container" width="600" border="0" cellspacing="0" cellpadding="0" style="background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 15px rgba(0,0,0,0.05);">
                <tr><td height="6" style="background-color: #007BFF;"></td></tr>
                <tr>
                    <td align="center" style="padding: 40px 0 20px 0;">
                        <img src="https://cdn.ppob.id/server/logo.png" alt="Logo PPOB.id" width="150" style="display: block;">
                    </td>
                </tr>
                <tr>
                    <td class="separator-padding" style="padding: 0 50px;">
                        <div style="border-top: 1px solid #E9ECEF; height: 1px; line-height: 1px; width: 100%;"></div>
                    </td>
                </tr>
                <tr>
                    <td class="content-padding" style="padding: 20px 50px;">
                        <h1 style="color: #004085; font-size: 24px; margin-bottom: 20px; text-align: center;">
                            Alamat Email Berhasil Diubah
                        </h1>
                        <p style="font-size: 16px; line-height: 1.6; color: #555555; text-align: center; margin-bottom: 25px;">
                            Halo, <strong>{{USER_NAME}}</strong><br><br>
                            Email akun ppob.id Anda berhasil diubah.
                        </p>
                        <table width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #F8F9FA; border-radius: 8px; border: 1px solid #E9ECEF; margin-bottom: 25px;">
                            <tr>
                                <td style="padding: 15px 20px;">
                                    <table width="100%" border="0" cellspacing="0" cellpadding="0">
                                        <tr>
                                            <td style="font-size: 13px; color: #999999; padding-bottom: 5px;">Email Lama</td>
                                            <td align="right" style="font-size: 13px; color: #333333; font-weight: bold;">{{OLD_EMAIL}}</td>
                                        </tr>
                                        <tr>
                                            <td style="font-size: 13px; color: #999999; padding-bottom: 5px;">Email Baru</td>
                                            <td align="right" style="font-size: 13px; color: #333333; font-weight: bold;">{{NEW_EMAIL}}</td>
                                        </tr>
                                        <tr>
                                            <td style="font-size: 13px; color: #999999; padding-bottom: 5px;">Waktu Perubahan</td>
                                            <td align="right" style="font-size: 13px; color: #333333; font-weight: bold;">{{CHANGE_TIME}}</td>
                                        </tr>
                                    </table>
                                </td>
                            </tr>
                        </table>
                        <p style="font-size: 14px; line-height: 1.5; color: #333333; text-align: center; margin-bottom: 20px;">
                            Jika ini <strong>bukan Anda</strong>, segera amankan akun Anda.
                        </p>
                    </td>
                </tr>
                <tr>
                    <td class="content-padding" style="padding: 0 50px 40px 50px;">
                        <table width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #EBF5FF; border-radius: 8px; border: 1px solid #D6E9FF;">
                            <tr>
                                <td align="center" style="padding: 20px;">
                                    <p style="margin: 0; font-size: 14px; color: #004085; font-weight: bold;">Butuh Bantuan?</p>
                                    <p style="margin: 10px 0 0 0; font-size: 14px; color: #333333;">
                                        📧 <a href="mailto:cs@ppob.id" style="color: #007BFF; text-decoration: none;">cs@ppob.id</a> &nbsp;|&nbsp;
                                        📱 <a href="https://wa.me/628138181640" style="color: #007BFF; text-decoration: none;">+62 813-8181-640</a>
                                    </p>
                                </td>
                            </tr>
                        </table>
                    </td>
                </tr>
                <tr>
                    <td class="separator-padding" style="padding: 0 50px;">
                        <div style="border-top: 1px solid #E9ECEF; height: 1px; line-height: 1px; width: 100%;"></div>
                    </td>
                </tr>
                <tr>
                    <td style="padding: 35px 40px; color: #333333; text-align: center;">
                        <p style="margin: 0; font-weight: bold; font-size: 15px; letter-spacing: 0.5px; color: #004085;">PT Gerbang Transaksi Digital</p>
                        <p style="margin: 8px 0 25px 0; font-size: 12px; color: #666666; line-height: 1.5;">
                            Wisma KEIAI 14th Floor Unit 1410<br>
                            Jln. Jenderal Sudirman Kav.3, Karet Tengsin, Tanah Abang Jakarta Pusat, Jakarta, 10220
                        </p>
                        <p style="margin: 0 0 30px 0;">
                            <a href="https://facebook.com/ppob.id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-facebook.svg" alt="Facebook" width="28" height="28" style="display: block; border: 0;">
                            </a>
                            <a href="https://instagram.com/ppob.id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-instagram.svg" alt="Instagram" width="28" height="28" style="display: block; border: 0;">
                            </a>
                            <a href="https://twitter.com/ppob_id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-x.svg" alt="Twitter X" width="28" height="28" style="display: block; border: 0;">
                            </a>
                            <a href="https://tiktok.com/@ppob.id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-tiktok.svg" alt="TikTok" width="28" height="28" style="display: block; border: 0;">
                            </a>
                        </p>
                        <table width="100%" border="0" cellspacing="0" cellpadding="0">
                            <tr>
                                <td align="center">
                                    <p style="font-size: 13px; font-weight: bold; color: #333333; margin: 0 0 15px 0; letter-spacing: 0.5px;">
                                        Unduh ppob.id melalui
                                    </p>
                                </td>
                            </tr>
                            <tr>
                                <td align="center" style="padding-bottom: 25px;">
                                    <table border="0" cellspacing="0" cellpadding="0" style="margin: 0 auto;">
                                        <tr>
                                            <td valign="middle" style="padding-right: 15px;">
                                                <img src="https://cdn.ppob.id/server/icon.png" alt="Icon" width="42" height="42" style="display: block; border-radius: 8px;">
                                            </td>
                                            <td valign="middle" style="padding-right: 8px;">
                                                <a href="#">
                                                    <img src="https://cdn.ppob.id/server/google-play.png" alt="Google Play" height="35" style="display: block; border:0;" class="store-btn">
                                                </a>
                                            </td>
                                            <td valign="middle">
                                                <a href="#">
                                                    <img src="https://cdn.ppob.id/server/app-store.png" alt="App Store" height="35" style="display: block; border:0;" class="store-btn">
                                                </a>
                                            </td>
                                        </tr>
                                    </table>
                                </td>
                            </tr>
                        </table>
                        <p style="margin: 0; font-size: 11px; color: #999999;">
                            &copy; 2026 PT Gerbang Transaksi Digital. All rights reserved.
                        </p>
                    </td>
                </tr>
            </table>
            <table width="100%" border="0" cellspacing="0" cellpadding="0">
                <tr><td height="40"></td></tr>
            </table>
        </td>
    </tr>
</table>
</body>
</html>',
'Halo, {{USER_NAME}}

Email akun ppob.id Anda berhasil diubah.

Email Lama: {{OLD_EMAIL}}
Email Baru: {{NEW_EMAIL}}
Waktu Perubahan: {{CHANGE_TIME}}

Jika ini bukan Anda, segera amankan akun Anda.

ppob.id security',
4,
'[{"name": "USER_NAME", "type": "string", "required": true}, {"name": "OLD_EMAIL", "type": "string", "required": true}, {"name": "NEW_EMAIL", "type": "string", "required": true}, {"name": "CHANGE_TIME", "type": "string", "required": true}]',
TRUE, 90);

-- 4. Security Alert - Phone Changed
('emt_security_phone_number_changed', 'SECURITY_PHONE_NUMBER_CHANGED', 'Email Diubah', 'Notifikasi ketika nomor telepon diubah', 'SECURITY',
    'Nomor Telepon Berhasil Diubah',
    '<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Nomor Telepon Berhasil Diubah</title>
    <style>
        body { margin: 0; padding: 0; -webkit-text-size-adjust: 100%; -ms-text-size-adjust: 100%; background-color: #F0F4F8; }
        table { border-spacing: 0; border-collapse: collapse; }
        img { border: 0; line-height: 100%; outline: none; text-decoration: none; display: block; }

        body, td { font-family: ''Helvetica Neue'', Helvetica, Arial, sans-serif; color: #333333; }

        .btn-primary {
            background-color: #007BFF;
            color: #ffffff !important;
            padding: 12px 24px;
            border-radius: 6px;
            text-decoration: none;
            font-weight: bold;
            display: inline-block;
            mso-padding-alt: 0;
            text-align: center;
        }

        .footer-link:hover { opacity: 0.8 !important; }

        @media only screen and (max-width: 600px) {
            .container { width: 100% !important; padding: 0 !important; }
            .content-padding { padding: 20px !important; }
            .store-btn { width: 130px !important; height: auto !important; }
            .separator-padding { padding: 0 20px !important; }
        }
    </style>
</head>
<body style="background-color: #F0F4F8; padding: 40px 0;">
<table width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #F0F4F8;">
    <tr>
        <td align="center">
            <table class="container" width="600" border="0" cellspacing="0" cellpadding="0" style="background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 15px rgba(0,0,0,0.05);">
                <tr><td height="6" style="background-color: #007BFF;"></td></tr>
                <tr>
                    <td align="center" style="padding: 40px 0 20px 0;">
                        <img src="https://cdn.ppob.id/server/logo.png" alt="Logo PPOB.id" width="150" style="display: block;">
                    </td>
                </tr>
                <tr>
                    <td class="separator-padding" style="padding: 0 50px;">
                        <div style="border-top: 1px solid #E9ECEF; height: 1px; line-height: 1px; width: 100%;"></div>
                    </td>
                </tr>
                <tr>
                    <td class="content-padding" style="padding: 20px 50px;">
                        <h1 style="color: #004085; font-size: 24px; margin-bottom: 20px; text-align: center;">
                            Nomor Telepon Berhasil Diubah
                        </h1>
                        <p style="font-size: 16px; line-height: 1.6; color: #555555; text-align: center; margin-bottom: 25px;">
                            Halo, <strong>{{USER_NAME}}</strong><br><br>
                            Nomor Telepon akun ppob.id Anda berhasil diubah.
                        </p>
                        <table width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #F8F9FA; border-radius: 8px; border: 1px solid #E9ECEF; margin-bottom: 25px;">
                            <tr>
                                <td style="padding: 15px 20px;">
                                    <table width="100%" border="0" cellspacing="0" cellpadding="0">
                                        <tr>
                                            <td style="font-size: 13px; color: #999999; padding-bottom: 5px;">Nomor Lama</td>
                                            <td align="right" style="font-size: 13px; color: #333333; font-weight: bold;">{{OLD_PHONE_NUMBER}}</td>
                                        </tr>
                                        <tr>
                                            <td style="font-size: 13px; color: #999999; padding-bottom: 5px;">Nomor Baru</td>
                                            <td align="right" style="font-size: 13px; color: #333333; font-weight: bold;">{{NEW_PHONE_NUMBER}}</td>
                                        </tr>
                                        <tr>
                                            <td style="font-size: 13px; color: #999999; padding-bottom: 5px;">Waktu Perubahan</td>
                                            <td align="right" style="font-size: 13px; color: #333333; font-weight: bold;">{{CHANGE_TIME}}</td>
                                        </tr>
                                    </table>
                                </td>
                            </tr>
                        </table>
                        <p style="font-size: 14px; line-height: 1.5; color: #333333; text-align: center; margin-bottom: 20px;">
                            Jika ini <strong>bukan Anda</strong>, segera amankan akun Anda.
                        </p>
                    </td>
                </tr>
                <tr>
                    <td class="content-padding" style="padding: 0 50px 40px 50px;">
                        <table width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #EBF5FF; border-radius: 8px; border: 1px solid #D6E9FF;">
                            <tr>
                                <td align="center" style="padding: 20px;">
                                    <p style="margin: 0; font-size: 14px; color: #004085; font-weight: bold;">Butuh Bantuan?</p>
                                    <p style="margin: 10px 0 0 0; font-size: 14px; color: #333333;">
                                        📧 <a href="mailto:cs@ppob.id" style="color: #007BFF; text-decoration: none;">cs@ppob.id</a> &nbsp;|&nbsp;
                                        📱 <a href="https://wa.me/628138181640" style="color: #007BFF; text-decoration: none;">+62 813-8181-640</a>
                                    </p>
                                </td>
                            </tr>
                        </table>
                    </td>
                </tr>
                <tr>
                    <td class="separator-padding" style="padding: 0 50px;">
                        <div style="border-top: 1px solid #E9ECEF; height: 1px; line-height: 1px; width: 100%;"></div>
                    </td>
                </tr>
                <tr>
                    <td style="padding: 35px 40px; color: #333333; text-align: center;">
                        <p style="margin: 0; font-weight: bold; font-size: 15px; letter-spacing: 0.5px; color: #004085;">PT Gerbang Transaksi Digital</p>
                        <p style="margin: 8px 0 25px 0; font-size: 12px; color: #666666; line-height: 1.5;">
                            Wisma KEIAI 14th Floor Unit 1410<br>
                            Jln. Jenderal Sudirman Kav.3, Karet Tengsin, Tanah Abang Jakarta Pusat, Jakarta, 10220
                        </p>
                        <p style="margin: 0 0 30px 0;">
                            <a href="https://facebook.com/ppob.id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-facebook.svg" alt="Facebook" width="28" height="28" style="display: block; border: 0;">
                            </a>
                            <a href="https://instagram.com/ppob.id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-instagram.svg" alt="Instagram" width="28" height="28" style="display: block; border: 0;">
                            </a>
                            <a href="https://twitter.com/ppob_id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-x.svg" alt="Twitter X" width="28" height="28" style="display: block; border: 0;">
                            </a>
                            <a href="https://tiktok.com/@ppob.id" target="_blank" class="footer-link" style="text-decoration: none; margin: 0 8px; display: inline-block;">
                                <img src="https://cdn.ppob.id/server/social-tiktok.svg" alt="TikTok" width="28" height="28" style="display: block; border: 0;">
                            </a>
                        </p>
                        <table width="100%" border="0" cellspacing="0" cellpadding="0">
                            <tr>
                                <td align="center">
                                    <p style="font-size: 13px; font-weight: bold; color: #333333; margin: 0 0 15px 0; letter-spacing: 0.5px;">
                                        Unduh ppob.id melalui
                                    </p>
                                </td>
                            </tr>
                            <tr>
                                <td align="center" style="padding-bottom: 25px;">
                                    <table border="0" cellspacing="0" cellpadding="0" style="margin: 0 auto;">
                                        <tr>
                                            <td valign="middle" style="padding-right: 15px;">
                                                <img src="https://cdn.ppob.id/server/icon.png" alt="Icon" width="42" height="42" style="display: block; border-radius: 8px;">
                                            </td>
                                            <td valign="middle" style="padding-right: 8px;">
                                                <a href="#">
                                                    <img src="https://cdn.ppob.id/server/google-play.png" alt="Google Play" height="35" style="display: block; border:0;" class="store-btn">
                                                </a>
                                            </td>
                                            <td valign="middle">
                                                <a href="#">
                                                    <img src="https://cdn.ppob.id/server/app-store.png" alt="App Store" height="35" style="display: block; border:0;" class="store-btn">
                                                </a>
                                            </td>
                                        </tr>
                                    </table>
                                </td>
                            </tr>
                        </table>
                        <p style="margin: 0; font-size: 11px; color: #999999;">
                            &copy; 2026 PT Gerbang Transaksi Digital. All rights reserved.
                        </p>
                    </td>
                </tr>
            </table>
            <table width="100%" border="0" cellspacing="0" cellpadding="0">
                <tr><td height="40"></td></tr>
            </table>
        </td>
    </tr>
</table>
</body>
</html>',
    'Halo, {{USER_NAME}}

    Nomor Telepon akun ppob.id Anda berhasil diubah.

    Nomor Lama: {{OLD_PHONE_NUMBER}}
    Nomor Baru: {{NEW_PHONE_NUMBER}}
    Waktu Perubahan: {{CHANGE_TIME}}

    Jika ini bukan Anda, segera amankan akun Anda.

    ppob.id security',
    4,
    '[{"name": "USER_NAME", "type": "string", "required": true}, {"name": "OLD_PHONE_NUMBER", "type": "int", "required": true}, {"name": "NEW_PHONE_NUMBER", "type": "int", "required": true}, {"name": "CHANGE_TIME", "type": "string", "required": true}]',
    TRUE, 90);

-- Email send logs for tracking
CREATE TABLE IF NOT EXISTS email_logs (
    id VARCHAR(36) PRIMARY KEY,                           -- UUID: eml_xxx
    template_id VARCHAR(36) NOT NULL,                     -- FK to email_templates
    template_code VARCHAR(50) NOT NULL,                   -- Denormalized for quick access

    -- Recipient
    recipient_email VARCHAR(255) NOT NULL,
    recipient_name VARCHAR(100),
    user_id VARCHAR(36),                                  -- FK to users (nullable for non-user emails)

    -- Email Details
    subject VARCHAR(255) NOT NULL,
    params JSONB,                                         -- Parameters used for this email

    -- Provider
    provider VARCHAR(20) NOT NULL,                        -- BREVO, RESEND
    provider_message_id VARCHAR(255),                     -- Message ID from provider

    -- Status
    status VARCHAR(20) DEFAULT 'pending',                 -- pending, sent, delivered, failed, bounced
    error_message TEXT,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,

    CONSTRAINT fk_email_logs_template FOREIGN KEY (template_id) REFERENCES email_templates(id)
);

-- Indexes for email_logs
CREATE INDEX idx_email_logs_template_id ON email_logs(template_id);
CREATE INDEX idx_email_logs_template_code ON email_logs(template_code);
CREATE INDEX idx_email_logs_user_id ON email_logs(user_id);
CREATE INDEX idx_email_logs_status ON email_logs(status);
CREATE INDEX idx_email_logs_created_at ON email_logs(created_at);
CREATE INDEX idx_email_logs_provider ON email_logs(provider);
