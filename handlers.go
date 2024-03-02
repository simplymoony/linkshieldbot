package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/simplymoony/linkshieldbot/internal/tg"
	"github.com/simplymoony/linkshieldbot/internal/tg/poller"
)

// Routes updates to their appropriate handler.
func routeUpdate(e *env) poller.UpdateCallback {
	return func(ctx context.Context, bot *tg.Bot, update *tg.Update) error {
		switch {
		case update.Message != nil:
			if update.Message.Chat.Type == "private" &&
				update.Message.Text == "/start" {
				return onCMDStart(ctx, bot, update.Message, e)
			}
		case update.ChatJoinRequest != nil:
			return onJoinRequest(ctx, bot, update.ChatJoinRequest, e)
		}

		return nil
	}
}

// Handles errors that occur in handlers.
func handleError(e *env) poller.ErrorCallback {
	return func(err error, fromPoller bool) {
		if errors.Is(err, context.Canceled) {
			return
		}
		if fromPoller {
			e.Printf("Failed to fetch updates, retrying.. (%v)", err)
		} else {
			e.Printf("Failed to process update: %v", err)
		}
	}
}

// Handles command /start.
func onCMDStart(ctx context.Context, bot *tg.Bot, msg *tg.Message, e *env) error {
	e.Verbosef("Received /start command (user_id=%d)", msg.Chat.ID)

	const text = "Hey there! I'm a private instance of " +
		"<a href=\"https://t.me/LinkShieldBot\">LinkShieldBot</a> - " +
		"a bot to filter unwanted chat join requests.\n" +
		"Check my channel out to learn more or run your own instance."

	if _, err := bot.SendMessage(ctx, msg.Chat.ID, text, &tg.SendMessageOpts{
		ParseMode: "HTML",
	}); err != nil {
		return err
	}

	return nil
}

// Handles chat join requests.
func onJoinRequest(ctx context.Context, bot *tg.Bot, req *tg.ChatJoinRequest, e *env) error {
	src, ok := e.Directives[strconv.FormatInt(req.Chat.ID, 10)]
	if !ok {
		e.Verbosef(
			"Received join request but directive is missing for chat, skipping.. (chat_id=%d)",
			req.Chat.ID,
		)
		return nil
	}

	e.Verbosef("Received join request (chat_id=%d, user_id=%d)", req.Chat.ID, req.UserChatID)

	member, err := bot.GetChatMember(ctx, src, req.UserChatID)
	if err != nil {
		return fmt.Errorf(
			"failed to get chat member (chat_id=%d, user_id=%d): %v",
			src, req.UserChatID, err,
		)
	}

	if member.Status != "member" && member.Status != "creator" && member.Status != "administrator" &&
		member.Status != "restricted" {

		ok, err := bot.DeclineChatJoinRequest(ctx, req.Chat.ID, req.UserChatID)
		if err != nil {
			return fmt.Errorf(
				"failed to decline join request (chat_id=%d, user_id=%d): %v",
				req.Chat.ID, req.UserChatID, err,
			)
		}
		if !ok {
			e.Printf(
				"Couldn't decline user %d in chat %d, maybe already declined?",
				req.UserChatID, req.Chat.ID,
			)
		}
		e.Verbosef("Declined join request (chat_id=%d, user_id=%d)", req.Chat.ID, req.UserChatID)

		return nil
	}

	ok, err = bot.ApproveChatJoinRequest(ctx, req.Chat.ID, req.UserChatID)
	if err != nil {
		return fmt.Errorf(
			"failed to accept join request (chat_id=%d, user_id=%d): %v",
			req.Chat.ID, req.UserChatID, err,
		)
	}
	if !ok {
		e.Printf(
			"Couldn't approve user %d in chat %d, maybe already approved?",
			req.UserChatID, req.Chat.ID,
		)
	}
	e.Verbosef(
		"Accepted join request (chat_id=%d, user_id=%d)", req.Chat.ID, req.UserChatID)

	return nil
}
