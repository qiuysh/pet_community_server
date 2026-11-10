-- R5：宠物可见范围（0公开 / 1仅粉丝 / 2仅自己），默认公开
ALTER TABLE pets
  ADD COLUMN visibility TINYINT NOT NULL DEFAULT 0 COMMENT '可见范围：0公开 1仅粉丝 2仅自己'
  AFTER status;

CREATE INDEX idx_pets_visibility ON pets (visibility);
