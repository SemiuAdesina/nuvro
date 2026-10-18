INSERT INTO destinations (id, name, region, country_code, timezone, latitude, longitude, primary_airport, crowd_profile, pace_profile) VALUES
    ('zanzibar-tz', 'Zanzibar', 'Unguja', 'TZ', 'Africa/Dar_es_Salaam', -6.1659, 39.2026, 'ZNZ', 'quiet', 'relaxed'),
    ('marrakesh-ma', 'Marrakesh', 'Marrakesh-Safi', 'MA', 'Africa/Casablanca', 31.6295, -7.9811, 'RAK', 'lively', 'balanced'),
    ('cape-town-za', 'Cape Town', 'Western Cape', 'ZA', 'Africa/Johannesburg', -33.9249, 18.4241, 'CPT', 'any', 'balanced'),
    ('madeira-pt', 'Madeira', 'Madeira', 'PT', 'Atlantic/Madeira', 32.7607, -16.9595, 'FNC', 'quiet', 'relaxed'),
    ('barcelona-es', 'Barcelona', 'Catalonia', 'ES', 'Europe/Madrid', 41.3874, 2.1686, 'BCN', 'lively', 'packed'),
    ('reykjavik-is', 'Reykjavik', 'Capital Region', 'IS', 'Atlantic/Reykjavik', 64.1466, -21.9426, 'KEF', 'quiet', 'balanced'),
    ('phuket-th', 'Phuket', 'Phuket', 'TH', 'Asia/Bangkok', 7.8804, 98.3923, 'HKT', 'lively', 'relaxed'),
    ('bali-id', 'Bali', 'Bali', 'ID', 'Asia/Makassar', -8.4095, 115.1889, 'DPS', 'any', 'relaxed'),
    ('dubai-ae', 'Dubai', 'Dubai', 'AE', 'Asia/Dubai', 25.2048, 55.2708, 'DXB', 'lively', 'packed'),
    ('lisbon-pt', 'Lisbon', 'Lisbon', 'PT', 'Europe/Lisbon', 38.7223, -9.1393, 'LIS', 'any', 'balanced'),
    ('mexico-city-mx', 'Mexico City', 'Mexico City', 'MX', 'America/Mexico_City', 19.4326, -99.1332, 'MEX', 'lively', 'packed'),
    ('cartagena-co', 'Cartagena', 'Bolivar', 'CO', 'America/Bogota', 10.3910, -75.4794, 'CTG', 'any', 'relaxed'),
    ('bridgetown-bb', 'Bridgetown', 'Saint Michael', 'BB', 'America/Barbados', 13.1132, -59.5988, 'BGI', 'quiet', 'relaxed'),
    ('queenstown-nz', 'Queenstown', 'Otago', 'NZ', 'Pacific/Auckland', -45.0312, 168.6626, 'ZQN', 'quiet', 'packed')
ON CONFLICT (id) DO NOTHING;

-- +seed statement
INSERT INTO destination_tags (destination_id, kind, tag) VALUES
    ('zanzibar-tz', 'climate', 'warm'), ('zanzibar-tz', 'interest', 'beach'), ('zanzibar-tz', 'interest', 'food'), ('zanzibar-tz', 'interest', 'relaxation'), ('zanzibar-tz', 'style', 'island'),
    ('marrakesh-ma', 'climate', 'warm'), ('marrakesh-ma', 'interest', 'culture'), ('marrakesh-ma', 'interest', 'food'), ('marrakesh-ma', 'interest', 'shopping'), ('marrakesh-ma', 'style', 'city'),
    ('cape-town-za', 'climate', 'mild'), ('cape-town-za', 'interest', 'beach'), ('cape-town-za', 'interest', 'nature'), ('cape-town-za', 'interest', 'food'), ('cape-town-za', 'style', 'city'),
    ('madeira-pt', 'climate', 'mild'), ('madeira-pt', 'interest', 'nature'), ('madeira-pt', 'interest', 'relaxation'), ('madeira-pt', 'interest', 'beach'), ('madeira-pt', 'style', 'island'),
    ('barcelona-es', 'climate', 'warm'), ('barcelona-es', 'interest', 'culture'), ('barcelona-es', 'interest', 'food'), ('barcelona-es', 'interest', 'nightlife'), ('barcelona-es', 'style', 'city'),
    ('reykjavik-is', 'climate', 'cold'), ('reykjavik-is', 'interest', 'nature'), ('reykjavik-is', 'interest', 'adventure'), ('reykjavik-is', 'style', 'outdoors'),
    ('phuket-th', 'climate', 'warm'), ('phuket-th', 'interest', 'beach'), ('phuket-th', 'interest', 'food'), ('phuket-th', 'interest', 'relaxation'), ('phuket-th', 'style', 'island'),
    ('bali-id', 'climate', 'warm'), ('bali-id', 'interest', 'beach'), ('bali-id', 'interest', 'relaxation'), ('bali-id', 'interest', 'nature'), ('bali-id', 'style', 'island'),
    ('dubai-ae', 'climate', 'warm'), ('dubai-ae', 'interest', 'shopping'), ('dubai-ae', 'interest', 'food'), ('dubai-ae', 'interest', 'beach'), ('dubai-ae', 'style', 'city'),
    ('lisbon-pt', 'climate', 'mild'), ('lisbon-pt', 'interest', 'culture'), ('lisbon-pt', 'interest', 'food'), ('lisbon-pt', 'interest', 'beach'), ('lisbon-pt', 'style', 'city'),
    ('mexico-city-mx', 'climate', 'mild'), ('mexico-city-mx', 'interest', 'culture'), ('mexico-city-mx', 'interest', 'food'), ('mexico-city-mx', 'interest', 'history'), ('mexico-city-mx', 'style', 'city'),
    ('cartagena-co', 'climate', 'warm'), ('cartagena-co', 'interest', 'beach'), ('cartagena-co', 'interest', 'history'), ('cartagena-co', 'interest', 'food'), ('cartagena-co', 'style', 'coastal'),
    ('bridgetown-bb', 'climate', 'warm'), ('bridgetown-bb', 'interest', 'beach'), ('bridgetown-bb', 'interest', 'relaxation'), ('bridgetown-bb', 'style', 'island'),
    ('queenstown-nz', 'climate', 'cold'), ('queenstown-nz', 'interest', 'adventure'), ('queenstown-nz', 'interest', 'nature'), ('queenstown-nz', 'style', 'outdoors')
ON CONFLICT DO NOTHING;
