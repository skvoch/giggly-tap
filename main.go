package main

import (
	"fmt"
	"github.com/micmonay/keybd_event"
	"github.com/moutend/go-hook/pkg/keyboard"
	"github.com/moutend/go-hook/pkg/types"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"time"
)

var (
	targetKeys = map[types.VKCode]keybd_event.KeyBonding{
		types.VK_A: {},
		types.VK_S: {},
		types.VK_D: {},
	}
)

func IsTargetKey(key types.VKCode) bool {
	_, ok := targetKeys[key]
	return ok
}

func ReleaseAllOther(inputCode types.VKCode) {
	for code, key := range targetKeys {
		if code != inputCode {
			err := key.Release()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func main() {
	logger := log.Logger

	for key, _ := range targetKeys {
		bonding, err := keybd_event.NewKeyBonding()
		if err != nil {
			logger.Fatal().Err(err).Send()
		}

		bonding.SetKeys(int(key))
		targetKeys[key] = bonding
	}

	keyboardChan := make(chan types.KeyboardEvent, 100)

	if err := keyboard.Install(nil, keyboardChan); err != nil {
		logger.Fatal().Err(err).Send()
		return
	}

	defer keyboard.Uninstall()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	bonding := targetKeys[types.VK_A]
	err := bonding.Press()
	if err != nil {
		logger.Error().Err(err).Send()
	}

	logger.Info().Msg("start capturing keyboard input")

	for {
		select {
		case <-time.After(5 * time.Minute):
			fmt.Println("Received timeout signal")
			return
		case <-signalChan:
			fmt.Println("Received shutdown signal")
			return

		case event := <-keyboardChan:

			if IsTargetKey(event.VKCode) {

				logger.Error().Msg("pressed target key")
				if event.Message == types.WM_KEYDOWN {

					go ReleaseAllOther(event.VKCode)
				}
			}
		}
	}
}
