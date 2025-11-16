-- Создание фейковых пользователей
INSERT INTO users (id, login, password_hash, created_at) VALUES
    ('11111111-1111-1111-1111-111111111111', 'test_user', '$2a$10$dummyhash', NOW() - INTERVAL '30 days'),
    ('22222222-2222-2222-2222-222222222222', 'player_one', '$2a$10$dummyhash', NOW() - INTERVAL '15 days'),
    ('33333333-3333-3333-3333-333333333333', 'gamer_pro', '$2a$10$dummyhash', NOW() - INTERVAL '7 days')
ON CONFLICT (login) DO NOTHING;

-- Создание фейковых реплеев
-- Формат пути: storage/<год>/<месяц>/<id>.rep.gz
INSERT INTO replays (id, original_name, file_path, size_bytes, uploaded_at, compression, compressed, user_id) VALUES
    (
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'match_2024_01_15.rep',
        'storage/2024/01/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.rep.gz',
        1048576,
        '2024-01-15 14:30:00+00',
        'gzip',
        true,
        '11111111-1111-1111-1111-111111111111'
    ),
    (
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        'tournament_final.rep',
        'storage/2024/02/bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb.rep.gz',
        2097152,
        '2024-02-20 18:45:00+00',
        'gzip',
        true,
        '11111111-1111-1111-1111-111111111111'
    ),
    (
        'cccccccc-cccc-cccc-cccc-cccccccccccc',
        'quick_match.rep',
        'storage/2024/03/cccccccc-cccc-cccc-cccc-cccccccccccc.rep.gz',
        524288,
        '2024-03-10 10:15:00+00',
        'gzip',
        true,
        '22222222-2222-2222-2222-222222222222'
    ),
    (
        'dddddddd-dddd-dddd-dddd-dddddddddddd',
        'ranked_game.rep',
        'storage/2024/03/dddddddd-dddd-dddd-dddd-dddddddddddd.rep.gz',
        1572864,
        '2024-03-25 20:00:00+00',
        'gzip',
        true,
        '22222222-2222-2222-2222-222222222222'
    ),
    (
        'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
        'practice_session.rep',
        'storage/2024/04/eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee.rep.gz',
        786432,
        '2024-04-05 12:30:00+00',
        'gzip',
        true,
        '33333333-3333-3333-3333-333333333333'
    ),
    (
        'ffffffff-ffff-ffff-ffff-ffffffffffff',
        'championship_round.rep',
        'storage/2024/04/ffffffff-ffff-ffff-ffff-ffffffffffff.rep.gz',
        3145728,
        '2024-04-18 16:20:00+00',
        'gzip',
        true,
        '33333333-3333-3333-3333-333333333333'
    );
