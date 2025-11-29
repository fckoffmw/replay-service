INSERT INTO users (id, login, password_hash, created_at) VALUES
    ('00000000-0000-0000-0000-000000000001', 'test_user', '$2a$10$dummyhash', NOW() - INTERVAL '30 days'),
    ('11111111-1111-1111-1111-111111111111', 'player_one', '$2a$10$dummyhash', NOW() - INTERVAL '15 days'),
    ('22222222-2222-2222-2222-222222222222', 'gamer_pro', '$2a$10$dummyhash', NOW() - INTERVAL '7 days')
ON CONFLICT (login) DO NOTHING;

INSERT INTO games (id, name, user_id, created_at) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Counter-Strike 2', '00000000-0000-0000-0000-000000000001', NOW() - INTERVAL '20 days'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Dota 2', '00000000-0000-0000-0000-000000000001', NOW() - INTERVAL '15 days'),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'Valorant', '11111111-1111-1111-1111-111111111111', NOW() - INTERVAL '10 days')
ON CONFLICT (user_id, name) DO NOTHING;

INSERT INTO replays (id, title, original_name, file_path, size_bytes, uploaded_at, compression, compressed, comment, game_id, user_id) VALUES
    (
        '10000000-0000-0000-0000-000000000001',
        'Epic comeback',
        'match_2024_01_15.rep',
        '00000000-0000-0000-0000-000000000001/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/10000000-0000-0000-0000-000000000001.rep',
        1048576,
        NOW() - INTERVAL '5 days',
        'none',
        false,
        'Amazing clutch in overtime',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        '00000000-0000-0000-0000-000000000001'
    ),
    (
        '10000000-0000-0000-0000-000000000002',
        'Tournament final',
        'tournament_final.rep',
        '00000000-0000-0000-0000-000000000001/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/10000000-0000-0000-0000-000000000002.rep',
        2097152,
        NOW() - INTERVAL '3 days',
        'none',
        false,
        'Won the championship',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        '00000000-0000-0000-0000-000000000001'
    ),
    (
        '10000000-0000-0000-0000-000000000003',
        'Rampage game',
        'rampage.rep',
        '00000000-0000-0000-0000-000000000001/bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb/10000000-0000-0000-0000-000000000003.rep',
        1572864,
        NOW() - INTERVAL '2 days',
        'none',
        false,
        'Got rampage with Invoker',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        '00000000-0000-0000-0000-000000000001'
    );
