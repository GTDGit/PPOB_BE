-- Migration: 040_seed_services_and_providers
-- Description: Seed service_categories, services, operators, ewallet_providers, pdam_regions, tv_providers
-- Created: 2026-04-08
-- Note: These tables were created in migrations 007, 009, 011 but never populated.
--        Data is migrated from hardcoded Go structs in home_repository.go and product_repository.go.

-- ============================================
-- 1. Service Categories (3 rows)
-- ============================================
INSERT INTO service_categories (id, name, slug, sort_order, is_active) VALUES
    ('prepaid',  'Pra Bayar',  'prepaid',  1, true),
    ('postpaid', 'Pasca Bayar', 'postpaid', 2, true),
    ('finance',  'Keuangan',   'finance',  3, true)
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- 2. Services (14 rows)
-- ============================================
INSERT INTO services (id, category_id, name, icon, icon_url, route, status, badge, is_featured, sort_order) VALUES
    -- Featured prepaid
    ('pulsa',           'prepaid',  'Pulsa',            'pulsa',           'https://cdn.ppob.id/icons/pulsa.png',           '/services/pulsa',           'active', NULL,   true,  1),
    ('paket_data',      'prepaid',  'Paket Data',       'paket_data',      'https://cdn.ppob.id/icons/paket-data.png',      '/services/paket-data',      'active', 'PROMO', true,  2),
    ('token_pln',       'prepaid',  'Token PLN',        'token_pln',       'https://cdn.ppob.id/icons/token-pln.png',       '/services/token-pln',       'active', NULL,   true,  3),
    ('voucher_game',    'prepaid',  'Voucher Game',     'voucher_game',    'https://cdn.ppob.id/icons/voucher-game.png',    '/services/voucher-game',    'active', NULL,   true,  8),

    -- Featured postpaid
    ('tagihan_pln',     'postpaid', 'Tagihan PLN',      'tagihan_pln',     'https://cdn.ppob.id/icons/tagihan-pln.png',     '/services/tagihan-pln',     'active', NULL,   true,  4),
    ('pdam',            'postpaid', 'PDAM',             'pdam',            'https://cdn.ppob.id/icons/pdam.png',            '/services/pdam',            'active', NULL,   true,  5),
    ('bpjs',            'postpaid', 'BPJS',             'bpjs',            'https://cdn.ppob.id/icons/bpjs.png',            '/services/bpjs',            'active', NULL,   true,  9),
    ('telkom',          'postpaid', 'Telkom',           'telkom',          'https://cdn.ppob.id/icons/telkom.png',           '/services/telkom',          'active', NULL,   true,  10),

    -- Postpaid only (not featured)
    ('pulsa_pascabayar','postpaid', 'Pulsa Pascabayar', 'pulsa_pascabayar','https://cdn.ppob.id/icons/pulsa-pascabayar.png','/services/pulsa-pascabayar','active', NULL,   false, 1),
    ('tagihan_gas',     'postpaid', 'Tagihan Gas',      'tagihan_gas',     'https://cdn.ppob.id/icons/tagihan-gas.png',     '/services/tagihan-gas',     'active', NULL,   false, 6),
    ('pbb',             'postpaid', 'PBB',              'pbb',             'https://cdn.ppob.id/icons/pbb.png',             '/services/pbb',             'active', NULL,   false, 7),
    ('tv_kabel',        'postpaid', 'TV Kabel',         'tv_kabel',        'https://cdn.ppob.id/icons/tv-kabel.png',        '/services/tv-kabel',        'active', NULL,   false, 8),

    -- Featured finance
    ('ewallet',         'finance',  'E-Wallet',         'ewallet',         'https://cdn.ppob.id/icons/ewallet.png',         '/services/ewallet',         'active', NULL,   true,  6),
    ('transfer_bank',   'finance',  'Transfer Bank',    'transfer_bank',   'https://cdn.ppob.id/icons/transfer-bank.png',   '/services/transfer-bank',   'active', NULL,   true,  7)
ON CONFLICT (id) DO UPDATE SET
    category_id = EXCLUDED.category_id,
    name = EXCLUDED.name,
    icon = EXCLUDED.icon,
    icon_url = EXCLUDED.icon_url,
    route = EXCLUDED.route,
    status = EXCLUDED.status,
    badge = EXCLUDED.badge,
    is_featured = EXCLUDED.is_featured,
    sort_order = EXCLUDED.sort_order,
    updated_at = CURRENT_TIMESTAMP;

-- ============================================
-- 3. Operators (7 rows)
-- ============================================
INSERT INTO operators (id, name, prefixes, icon, icon_url, status, sort_order) VALUES
    ('telkomsel', 'Telkomsel',       '["0811","0812","0813","0821","0822","0823","0852","0853"]',                         'telkomsel', 'https://cdn.ppob.id/operators/telkomsel.png', 'active', 1),
    ('indosat',   'Indosat Ooredoo', '["0814","0815","0816","0855","0856","0857","0858"]',                               'indosat',   'https://cdn.ppob.id/operators/indosat.png',   'active', 2),
    ('xl',        'XL Axiata',       '["0817","0818","0819","0859","0877","0878"]',                                       'xl',        'https://cdn.ppob.id/operators/xl.png',        'active', 3),
    ('axis',      'Axis',            '["0831","0832","0833","0838"]',                                                     'axis',      'https://cdn.ppob.id/operators/axis.png',      'active', 4),
    ('three',     'Tri',             '["0895","0896","0897","0898","0899"]',                                              'three',     'https://cdn.ppob.id/operators/three.png',     'active', 5),
    ('smartfren', 'Smartfren',       '["0881","0882","0883","0884","0885","0886","0887","0888","0889"]',                  'smartfren', 'https://cdn.ppob.id/operators/smartfren.png', 'active', 6),
    ('byu',       'by.U',            '["0851"]',                                                                          'byu',       'https://cdn.ppob.id/operators/byu.png',       'active', 7)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    prefixes = EXCLUDED.prefixes,
    icon = EXCLUDED.icon,
    icon_url = EXCLUDED.icon_url,
    status = EXCLUDED.status,
    sort_order = EXCLUDED.sort_order,
    updated_at = CURRENT_TIMESTAMP;

-- ============================================
-- 4. E-Wallet Providers (5 rows)
-- ============================================
INSERT INTO ewallet_providers (id, name, icon, icon_url, input_label, input_placeholder, input_type, status, sort_order) VALUES
    ('gopay',     'GoPay',     'gopay',     'https://cdn.ppob.id/ewallet/gopay.png',     'Nomor HP GoPay',     '08xxxxxxxxxx', 'phone', 'active', 1),
    ('ovo',       'OVO',       'ovo',       'https://cdn.ppob.id/ewallet/ovo.png',       'Nomor HP OVO',       '08xxxxxxxxxx', 'phone', 'active', 2),
    ('dana',      'DANA',      'dana',      'https://cdn.ppob.id/ewallet/dana.png',      'Nomor HP DANA',      '08xxxxxxxxxx', 'phone', 'active', 3),
    ('shopeepay', 'ShopeePay', 'shopeepay', 'https://cdn.ppob.id/ewallet/shopeepay.png', 'Nomor HP ShopeePay', '08xxxxxxxxxx', 'phone', 'active', 4),
    ('linkaja',   'LinkAja',   'linkaja',   'https://cdn.ppob.id/ewallet/linkaja.png',   'Nomor HP LinkAja',   '08xxxxxxxxxx', 'phone', 'active', 5)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    icon = EXCLUDED.icon,
    icon_url = EXCLUDED.icon_url,
    input_label = EXCLUDED.input_label,
    input_placeholder = EXCLUDED.input_placeholder,
    input_type = EXCLUDED.input_type,
    status = EXCLUDED.status,
    sort_order = EXCLUDED.sort_order,
    updated_at = CURRENT_TIMESTAMP;

-- ============================================
-- 5. PDAM Regions (8 rows)
-- ============================================
INSERT INTO pdam_regions (id, name, province, status, sort_order) VALUES
    ('pdam_jakarta',   'PDAM DKI Jakarta',              'DKI Jakarta',      'active', 1),
    ('pdam_bandung',   'PDAM Tirta Wening Kota Bandung','Jawa Barat',       'active', 2),
    ('pdam_surabaya',  'PDAM Surya Sembada Surabaya',   'Jawa Timur',       'active', 3),
    ('pdam_semarang',  'PDAM Tirta Moedal Semarang',    'Jawa Tengah',      'active', 4),
    ('pdam_medan',     'PDAM Tirtanadi Medan',          'Sumatera Utara',   'active', 5),
    ('pdam_palembang', 'PDAM Tirta Musi Palembang',     'Sumatera Selatan', 'active', 6),
    ('pdam_makassar',  'PDAM Makassar',                 'Sulawesi Selatan', 'active', 7),
    ('pdam_denpasar',  'PDAM Denpasar',                 'Bali',             'active', 8)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    province = EXCLUDED.province,
    status = EXCLUDED.status,
    sort_order = EXCLUDED.sort_order,
    updated_at = CURRENT_TIMESTAMP;

-- ============================================
-- 6. TV Providers (6 rows)
-- ============================================
INSERT INTO tv_providers (id, name, icon, icon_url, input_label, status, sort_order) VALUES
    ('indovision',    'Indovision',    'indovision',    'https://cdn.ppob.id/tv/indovision.png',    'Nomor Pelanggan', 'active', 1),
    ('transvision',   'Transvision',   'transvision',   'https://cdn.ppob.id/tv/transvision.png',   'Nomor Pelanggan', 'active', 2),
    ('topas',         'Topas TV',      'topas',         'https://cdn.ppob.id/tv/topas.png',         'Nomor Pelanggan', 'active', 3),
    ('firstmedia',    'First Media',   'firstmedia',    'https://cdn.ppob.id/tv/firstmedia.png',    'Nomor Pelanggan', 'active', 4),
    ('k_vision',      'K-Vision',      'kvision',       'https://cdn.ppob.id/tv/kvision.png',       'Nomor Pelanggan', 'active', 5),
    ('nex_parabola',  'Nex Parabola',  'nexparabola',   'https://cdn.ppob.id/tv/nexparabola.png',   'Nomor Pelanggan', 'active', 6)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    icon = EXCLUDED.icon,
    icon_url = EXCLUDED.icon_url,
    input_label = EXCLUDED.input_label,
    status = EXCLUDED.status,
    sort_order = EXCLUDED.sort_order,
    updated_at = CURRENT_TIMESTAMP;
