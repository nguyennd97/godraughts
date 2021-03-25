package godraughts

import (
	"../../../matchmaker_server"
	"../../../matchmaker_server/api"
	"./evt"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const (
	WaitForFirstMove uint8 = iota
	Playing
	Stopped
)

type State struct {
	timeLeft     map[int64]int64
	state        uint8
	turn         int64
	board        *BitBoard
	moveGenUtil  *MoveGenUtil
	disconnected map[int64]bool
	moveCount    int
}

const (
	StartGame        = "1"
	FirstMoveTimeout = "2"
	PlayerTimeout    = "3"
)

type CheckerPlayer struct {
	SID         uuid.UUID
	ID          int64
	Username    string
	DisplayName string
	Avatar      string
	Attributes  map[string]string
}

// implement server.RoomHandler
// @see: server.RoomHandler
type CheckersInternational struct {
	mux         sync.Mutex
	stopped     *atomic.Bool
	logger      *zap.Logger
	initTime    int64
	paddingTime int64
	players     []*CheckerPlayer
	// save game state
	state *State
}

func NewInternationalCheckers() *CheckersInternational {
	return &CheckersInternational{
		stopped: atomic.NewBool(false),
		players: []*CheckerPlayer{},
		state: &State{
			timeLeft:     make(map[int64]int64),
			state:        Stopped,
			board:        NewBitBoard(),
			moveGenUtil:  NewMoveGenUtil(),
			disconnected: make(map[int64]bool),
			moveCount:    0,
		},
	}
}

// @override
func (c *CheckersInternational) OnInit(room matchmaker_server.Room) {
	var err error
	if c.initTime, err = strconv.ParseInt(room.MetaData()["init_time"], 10, 64); err != nil {
		c.logger.Error("can not parse init_time", zap.Error(err))
	}

	if c.paddingTime, err = strconv.ParseInt(room.MetaData()["padding_time"], 10, 64); err != nil {
		c.logger.Error("can not parse padding_time", zap.Error(err))
	}

	c.logger = matchmaker_server.NewLogger(room.ID().String())

}

// @override
func (c *CheckersInternational) Processor(room matchmaker_server.Room, action string, data map[string]interface{}) {
	if c.stopped.Load() {
		return
	}

	c.mux.Lock()
	defer c.mux.Unlock()
	if action == StartGame {
		c.startGame(room)
	} else if action == FirstMoveTimeout {
		c.firstMoveTimeout(room)
	}
}

// @override
func (c *CheckersInternational) HandleJoin(room matchmaker_server.Room, user *matchmaker_server.User) {
	if len(c.players) == room.Max() {
		if player := c.getPlayerByID(user.ID); player != nil {
			// reconnect
			c.state.disconnected[user.ID] = false
			player.SID = user.SID // update sid
			c.sendAllRoomMessage(room, &evt.Evt{
				Event: &evt.Evt_PlayerReconnected{
					PlayerReconnected: &evt.PlayerReconnected{
						Id: user.ID,
					},
				},
			})
		} else {
			// ham allow join bi loi
			c.logger.Error("other player joined room when it's started")
		}
		return
	}

	// add nguoi choi moi
	c.players = append(c.players, &CheckerPlayer{
		SID:         user.SID,
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		Attributes:  user.Attributes,
	})
	c.state.disconnected[user.ID] = false

	// nếu đủ hai người chơi thì cho bắt đầu trận đấu
	if len(room.Players()) == room.Max() {
		if err := room.GetScheduler().Schedule(StartGame, nil, 0*time.Second); err != nil {
			c.logger.Error("schedule start game error", zap.Error(err))
		}
	}
}

// @override
func (c *CheckersInternational) HandleLeave(room matchmaker_server.Room, user *matchmaker_server.User) {
	if player := c.getPlayerByID(user.ID); player == nil {
		c.logger.Error("room leave error, conflict user id, can not find user id", zap.Int64("user id", user.ID))
		return
	} else {
		if c.state.state == WaitForFirstMove || c.state.state == Playing {
			c.state.disconnected[user.ID] = true
			c.sendAllRoomMessage(room, &evt.Evt{
				Event: &evt.Evt_PlayerDisconnected{
					PlayerDisconnected: &evt.PlayerDisconnected{
						Id: user.ID,
					}}})
		} else {
			c.removePlayer(user.ID)
			if len(c.players) == 0 {
				// hủy bàn
				room.CloseRoom()
			} else {
				c.sendAllRoomMessage(room, &evt.Evt{
					Event: &evt.Evt_PlayerLeft{
						PlayerLeft: &evt.PlayerLeft{
							Id: user.ID,
						}}})
			}
		}
	}

}

// @override
func (c *CheckersInternational) AllowJoin(room matchmaker_server.Room, user *matchmaker_server.User) *matchmaker_server.CheckJoinConditionResult {
	if len(c.players) == 2 {
		if disconnected, ok := c.state.disconnected[user.ID]; ok {
			if disconnected {
				return &matchmaker_server.CheckJoinConditionResult{
					Allow:  true,
					Reason: "",
				}
			}
		}
		return &matchmaker_server.CheckJoinConditionResult{
			Allow:  false,
			Reason: "room is full",
		}
	}

	if player := c.getPlayerByID(user.ID); player == nil {
		if c.state.state == WaitForFirstMove || c.state.state == Playing {
			return &matchmaker_server.CheckJoinConditionResult{
				Allow:  false,
				Reason: "game is started",
			}
		}
	}

	return &matchmaker_server.CheckJoinConditionResult{
		Allow:  true,
		Reason: "",
	}
}

// @override
func (c *CheckersInternational) HandleData(room matchmaker_server.Room, message *matchmaker_server.RoomDataMessage) {
	m := &evt.Evt{}

	if err := proto.Unmarshal(message.Data, m); err != nil {
		c.logger.Error(err.Error())
		return
	}

	switch m.Event.(type) {
	case *evt.Evt_DoMoveReq:
		c.playerMove(room, c.getPlayerBySid(message.From), m.GetDoMoveReq().Move)
	default:
		return
	}

}

// @override
func (c *CheckersInternational) OnClose(room matchmaker_server.Room) {
	if !c.stopped.CAS(false, true) {
		return
	}
}

func (c *CheckersInternational) startGame(room matchmaker_server.Room) {
	if c.state.state != Stopped {
		c.logger.Error("can not start game while it is started")
		return
	}
	c.state.board = NewBitBoard()
	c.state.moveGenUtil.GenerateMoves(c.state.board)
	c.state.turn = room.Players()[rand.Intn(2)].ID
	c.state.state = WaitForFirstMove

	for _, player := range c.players {
		c.state.timeLeft[player.ID] = c.initTime * 60
	}

	rand.Seed(time.Now().UTC().UnixNano())
	c.sendAllRoomMessage(room, &evt.Evt{Event: &evt.Evt_GameStarted{GameStarted: &evt.GameStarted{
		First: c.state.turn,
	}}})

	c.scheduleWithLogger(room, FirstMoveTimeout, nil, 30*time.Second)
}

func (c *CheckersInternational) firstMoveTimeout(room matchmaker_server.Room) {
	c.endGame(room, c.getOtherPlayerID(c.state.turn))
}

func (c *CheckersInternational) timeOut(room matchmaker_server.Room, id int64) {

}

func (c *CheckersInternational) playerMove(room matchmaker_server.Room, player *CheckerPlayer, move int64) {
	// nếu hết game thì không cho di chuyển quân
	if c.state.state == Stopped {
		return
	}
	// nếu player = nil thì là lỗi code
	if player == nil {
		c.logger.Error("player Move error: player nil")
		return
	}

	// nếu người chơi không có lượt
	if c.state.turn != player.ID {
		c.sendRoomMessage(room, player.SID, &evt.Evt{Event: &evt.Evt_ErrorMessage{
			ErrorMessage: &evt.ErrorMessage{
				Err: "not your turn",
			},
		}})
		return
	}

	// nếu nước đi không hợp lệ
	if !c.state.moveGenUtil.HasMove(move) {
		c.sendRoomMessage(room, player.SID, &evt.Evt{Event: &evt.Evt_DoMoveRes{
			DoMoveRes: &evt.DoMoveRes{
				Id:     player.ID,
				Move:   move,
				Accept: false,
			},
		}})
		return
	}

	c.state.board.doMove(move)
	c.state.moveCount++
	c.state.turn = c.getOtherPlayerID(player.ID)
	timeLeft := c.state.timeLeft[c.state.turn]

	if c.state.state == WaitForFirstMove {
		room.GetScheduler().CancelIfExist(FirstMoveTimeout)
		if c.state.moveCount == 1 {
			timeLeft = 30
			c.scheduleWithLogger(room, FirstMoveTimeout, nil, time.Duration(timeLeft)*time.Second)
		} else {
			c.state.state = Playing
		}
	}

	c.sendAllRoomMessage(room, &evt.Evt{Event: &evt.Evt_PlayerMove{PlayerMove: &evt.PlayerMove{
		Id:       player.ID,
		Move:     move,
		Next:     c.getOtherPlayerID(player.ID),
		TimeLeft: timeLeft,
	}}})

	c.state.moveGenUtil.GenerateMoves(c.state.board)
	c.logger.Debug(fmt.Sprintf("%d", len(c.state.moveGenUtil.moves)))
	if c.state.moveGenUtil.IsEndGame() {
		c.endGame(room, player.ID)
	} else {

	}

}

func (c *CheckersInternational) endGame(room matchmaker_server.Room, winner int64) {
	if c.state.state == Stopped {
		c.logger.Error("can not end game while game state is stopped")
		return
	}

	room.GetScheduler().CancelAll()

	c.state.state = Stopped

	if len(room.Players()) <= 0 {
		room.CloseRoom()
		return
	}
	c.sendAllRoomMessage(room, &evt.Evt{Event: &evt.Evt_EndGame{EndGame: &evt.EndGame{
		Winner: winner,
	}}})
}

/*
	utility functions
*/

func (c *CheckersInternational) sendAllRoomMessage(room matchmaker_server.Room, evt *evt.Evt) {
	if len(room.Players()) == 0 {
		return
	}

	if data, err := proto.Marshal(evt); err == nil {
		room.SendAll(&api.Packet{
			Message: &api.Packet_RoomMessage{
				RoomMessage: &api.RoomMessage{
					RoomType: room.Type(),
					RoomId:   room.ID().String(),
					From:     "server",
					Data:     data,
					Code:     -1,
					Time:     time.Now().Unix(),
				},
			},
		})
	} else {
		c.logger.Error(err.Error())
	}

}

func (c *CheckersInternational) sendRoomMessage(room matchmaker_server.Room, id uuid.UUID, evt *evt.Evt) {
	if data, err := proto.Marshal(evt); err == nil {
		room.Send(id, &api.Packet{
			Message: &api.Packet_RoomMessage{
				RoomMessage: &api.RoomMessage{
					RoomType: room.Type(),
					RoomId:   room.ID().String(),
					From:     "server",
					Data:     data,
					Code:     -1,
					Time:     time.Now().Unix(),
				},
			},
		})
	} else {
		c.logger.Error(err.Error())
	}
}

func (c *CheckersInternational) getPlayerBySid(sid uuid.UUID) *CheckerPlayer {
	for _, v := range c.players {
		if sid == v.SID {
			return v
		}
	}
	return nil
}

func (c *CheckersInternational) getPlayerByID(id int64) *CheckerPlayer {
	for i := 0; i < len(c.players); i++ {
		if c.players[i].ID == id {
			return c.players[i]
		}
	}
	return nil
}

func (c *CheckersInternational) getOtherPlayerID(id int64) int64 {
	for _, v := range c.players {
		if v.ID != id {
			return v.ID
		}
	}
	return 0
}

func (c *CheckersInternational) removePlayer(id int64) {
	for i := 0; i < len(c.players); i++ {
		if c.players[i].ID == id {
			c.players[i] = c.players[len(c.players)-1]
			c.players = c.players[:len(c.players)-1]
			return
		}
	}
	delete(c.state.disconnected, id)
	delete(c.state.timeLeft, id)
}

func (c *CheckersInternational) scheduleWithLogger(room matchmaker_server.Room, action string, data map[string]interface{}, time time.Duration) {
	if err := room.GetScheduler().Schedule(action, data, time); err != nil {
		c.logger.Error(err.Error())
	}
}
