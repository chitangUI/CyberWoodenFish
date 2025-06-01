-- 初始化数据库脚本
-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 创建索引以提高性能
-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_highest_score ON users(highest_score DESC);
CREATE INDEX IF NOT EXISTS idx_users_last_played_at ON users(last_played_at DESC);

-- 分数表索引
CREATE INDEX IF NOT EXISTS idx_scores_user_id ON scores(user_id);
CREATE INDEX IF NOT EXISTS idx_scores_score ON scores(score DESC);
CREATE INDEX IF NOT EXISTS idx_scores_created_at ON scores(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_scores_game_mode ON scores(game_mode);

-- 复合索引
CREATE INDEX IF NOT EXISTS idx_scores_user_score ON scores(user_id, score DESC);
CREATE INDEX IF NOT EXISTS idx_scores_mode_score ON scores(game_mode, score DESC);

-- 刷新令牌表索引
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- 创建一些示例数据 (可选)
-- INSERT INTO users (username, email, nickname, password_hash, highest_score, games_played)
-- VALUES 
--     ('demo_user', 'demo@muyu.com', 'Demo Player', '$2a$10$demo_hash', 1000, 5),
--     ('test_user', 'test@muyu.com', 'Test Player', '$2a$10$test_hash', 800, 3);

-- 创建数据库函数
-- 获取用户排名的函数
CREATE OR REPLACE FUNCTION get_user_rank(user_id_param INTEGER)
RETURNS INTEGER AS $$
DECLARE
    user_rank INTEGER;
BEGIN
    SELECT rank_number INTO user_rank
    FROM (
        SELECT id, ROW_NUMBER() OVER (ORDER BY highest_score DESC) as rank_number
        FROM users 
        WHERE highest_score > 0
    ) ranked_users
    WHERE id = user_id_param;
    
    RETURN COALESCE(user_rank, 0);
END;
$$ LANGUAGE plpgsql;

-- 更新用户统计信息的函数
CREATE OR REPLACE FUNCTION update_user_stats()
RETURNS TRIGGER AS $$
BEGIN
    -- 更新用户的最高分数和游戏次数
    UPDATE users 
    SET 
        highest_score = GREATEST(highest_score, NEW.score),
        games_played = games_played + 1,
        last_played_at = NEW.created_at
    WHERE id = NEW.user_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器，在插入新分数时自动更新用户统计
DROP TRIGGER IF EXISTS trigger_update_user_stats ON scores;
CREATE TRIGGER trigger_update_user_stats
    AFTER INSERT ON scores
    FOR EACH ROW
    EXECUTE FUNCTION update_user_stats();

-- 清理过期刷新令牌的函数
CREATE OR REPLACE FUNCTION cleanup_expired_refresh_tokens()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM refresh_tokens 
    WHERE expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;
