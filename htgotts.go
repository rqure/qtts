package qtts

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/rqure/qtts/handlers"
)

/**
 * Required:
 * - mplayer
 *
 * Use:
 *
 * speech := htgotts.Speech{Folder: "audio", Language: "en", Handler: MPlayer}
 */

// Speech struct
type Speech struct {
	Folder   string
	Language string
	Voice    string
	Proxy    string
	Handler  handlers.PlayerInterface
}

// Creates a speech file with a given name
func (speech *Speech) CreateSpeechFile(text string, fileName string) (string, error) {
	var err error

	f := speech.Folder + fileName + ".mp3"

	if err = speech.downloadIfNotExists(f, text); err != nil {
		return "", err
	}

	return f, nil
}

// Plays an existent .mp3 file
func (speech *Speech) PlaySpeechFile(fileName string) error {
	if speech.Handler == nil {
		mplayer := handlers.MPlayer{}
		return mplayer.Play(fileName)
	}

	return speech.Handler.Play(fileName)
}

// Speak downloads speech and plays it using mplayer
func (speech *Speech) Speak(text string) error {

	var err error
	generatedHashName := speech.generateHashName(text)

	fileName, err := speech.CreateSpeechFile(text, generatedHashName)
	if err != nil {
		return err
	}

	return speech.PlaySpeechFile(fileName)
}

/**
 * Download the voice file if does not exists.
 */
func (speech *Speech) downloadIfNotExists(fileName string, text string) error {
	// Check if the file already exists
	if _, err := os.Stat(fileName); err == nil {
		return nil
	}

	// Create a client for Google Text-to-Speech
	ctx := context.Background()
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create TTS client: %w", err)
	}
	defer client.Close()

	// Set up the request
	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: speech.Language,
			SsmlGender:   parseGender(speech.Voice), // Parse the gender
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	// Perform the TTS request
	resp, err := client.SynthesizeSpeech(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to synthesize speech: %w", err)
	}

	// Write the audio content to the specified file
	output, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer output.Close()

	if _, err = output.Write(resp.AudioContent); err != nil {
		return fmt.Errorf("failed to write audio content to file: %w", err)
	}

	return nil
}

func (speech *Speech) generateHashName(name string) string {
	hash := md5.Sum([]byte(name))
	return fmt.Sprintf("%s_%s", speech.Language, hex.EncodeToString(hash[:]))
}

// Helper function to parse gender from string to SSMLGender enum
func parseGender(gender string) texttospeechpb.SsmlVoiceGender {
	gender = strings.ToUpper(gender)
	switch gender {
	case "MALE":
		return texttospeechpb.SsmlVoiceGender_MALE
	case "FEMALE":
		return texttospeechpb.SsmlVoiceGender_FEMALE
	default:
		return texttospeechpb.SsmlVoiceGender_NEUTRAL
	}
}
