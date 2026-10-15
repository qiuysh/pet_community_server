package cache

import (
	"fmt"
	"time"

	"pet-community/config"
)

const relationTTL = 15 * time.Minute

func keyFollowEdge(followerID, followingID uint) string {
	return fmt.Sprintf("rel:follow:%d:%d", followerID, followingID)
}

func keyBlockEdge(blockerID, blockedID uint) string {
	return fmt.Sprintf("rel:block:%d:%d", blockerID, blockedID)
}

func keyBlockEither(a, b uint) string {
	if a > b {
		a, b = b, a
	}
	return fmt.Sprintf("rel:block_either:%d:%d", a, b)
}

func keyBlockIDs(userID uint) string {
	return fmt.Sprintf("rel:block_ids:%d", userID)
}

func getBoolEdge(key string) (exists bool, hit bool) {
	if !config.RedisEnabled() {
		return false, false
	}
	val, err := config.RDB.Get(ctxBG(), key).Result()
	if err != nil {
		return false, false
	}
	return val == "1", true
}

func setBoolEdge(key string, exists bool) {
	if !config.RedisEnabled() {
		return
	}
	v := "0"
	if exists {
		v = "1"
	}
	_ = config.RDB.Set(ctxBG(), key, v, relationTTL).Err()
}

// Follow edge
func GetFollowEdge(followerID, followingID uint) (exists bool, hit bool) {
	return getBoolEdge(keyFollowEdge(followerID, followingID))
}

func SetFollowEdge(followerID, followingID uint, exists bool) {
	setBoolEdge(keyFollowEdge(followerID, followingID), exists)
}

func InvalidateFollowEdge(followerID, followingID uint) {
	DeleteKeys(keyFollowEdge(followerID, followingID))
}

// Block edge / either
func GetBlockEdge(blockerID, blockedID uint) (exists bool, hit bool) {
	return getBoolEdge(keyBlockEdge(blockerID, blockedID))
}

func SetBlockEdge(blockerID, blockedID uint, exists bool) {
	setBoolEdge(keyBlockEdge(blockerID, blockedID), exists)
	DeleteKeys(keyBlockEither(blockerID, blockedID), keyBlockIDs(blockerID), keyBlockIDs(blockedID))
}

func GetBlockEither(a, b uint) (exists bool, hit bool) {
	return getBoolEdge(keyBlockEither(a, b))
}

func SetBlockEither(a, b uint, exists bool) {
	setBoolEdge(keyBlockEither(a, b), exists)
}

func InvalidateBlockRelation(userA, userB uint) {
	DeleteKeys(
		keyBlockEdge(userA, userB),
		keyBlockEdge(userB, userA),
		keyBlockEither(userA, userB),
		keyBlockIDs(userA),
		keyBlockIDs(userB),
	)
}

// Block ID list
func GetBlockRelationIDs(userID uint) ([]uint, bool) {
	var ids []uint
	if !GetJSON(keyBlockIDs(userID), &ids) {
		return nil, false
	}
	return ids, true
}

func SetBlockRelationIDs(userID uint, ids []uint) {
	if ids == nil {
		ids = []uint{}
	}
	SetJSON(keyBlockIDs(userID), ids, relationTTL)
}
