package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// Add the StringOrInt type and its UnmarshalJSON method
type StringOrInt string

var (
	inputActiveElement  *Element // Currently active input element
	inputTextBuffer     string   // Buffer for text input
	cursorBlinkTimer    uint32   // For cursor animation
	currentSceneIndex   int
	selectedButtonIndex int
	keyboard            [][]string
)

var menuButtonRects = make(map[int]sdl.Rect) // Scene index → hitbox

type Config struct {
	Title       string        `json:"title"`
	Author      string        `json:"author"`
	Description string        `json:"description"`
	Variables   Variables     `json:"variables"`
	Scenes      []SceneConfig `json:"scenes"`
}

type Variables struct {
	ButtonColor struct {
		R int `json:"r"`
		G int `json:"g"`
		B int `json:"b"`
	} `json:"buttonColor"`
	LabelColor struct {
		R int `json:"r"`
		G int `json:"g"`
		B int `json:"b"`
	} `json:"labelColor"`
	BackgroundImage string            `json:"backgroundImage"`
	Fonts           map[string]string `json:"fonts"`
	FontSizes       map[string]int    `json:"fontSizes"`
	Custom          map[string]interface{}
}

type SceneConfig struct {
	Name     string    `json:"name"`
	Elements []Element `json:"elements"`
}

type Element struct {
	Type          string      `json:"type"`
	Text          string      `json:"text"`
	Color         string      `json:"color"`
	X             int32       `json:"x"`
	Y             int32       `json:"y"`
	Font          string      `json:"font"`
	BgColor       string      `json:"bgColor"`
	Trigger       string      `json:"trigger"`
	TriggerTarget string      `json:"triggerTarget"`
	TriggerValue  string      `json:"triggerValue"`
	Image         string      `json:"image"`
	Width         StringOrInt `json:"width"`
	Height        StringOrInt `json:"height"`
	Video         string      `json:"video"`
	Variable      string      `json:"variable"`
	Command       string      `json:"command"`      // For collapsed list execution
	ListVariable  string      `json:"listVariable"` // For storing list data
}

type CollapsedListItem struct {
	Title       string `json:"title"`
	Header      string `json:"header"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

func executeCommandAndParse(config *Config, command string, listVariable string) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error executing command: %v", err)
		return
	}

	var items []CollapsedListItem
	if err := json.Unmarshal(output, &items); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return
	}

	// Store the parsed items in a variable
	config.Variables.Custom[listVariable] = items
}

var videoPlayed = false // rack if video has been played
var inputText string    // Global variable to store input text
var keyboardPosX, keyboardPosY int
var virtualKeyboardActive = false

// Add this UnmarshalJSON method for StringOrInt
func (s *StringOrInt) UnmarshalJSON(data []byte) error {
	// Try string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = StringOrInt(str)
		return nil
	}

	// Then try number
	var num int
	if err := json.Unmarshal(data, &num); err == nil {
		*s = StringOrInt(strconv.Itoa(num))
		return nil
	}

	// Handle null values
	if string(data) == "null" {
		*s = ""
		return nil
	}

	return fmt.Errorf("StringOrInt: expected string or integer, got %q", data)
}

func (v *Variables) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Handle known fields explicitly
	knownFields := map[string]bool{
		"buttonColor":     true,
		"labelColor":      true,
		"backgroundImage": true,
		"fonts":           true,
		"fontSizes":       true,
	}

	v.Custom = make(map[string]interface{})

	for key, val := range raw {
		if knownFields[key] {
			switch key {
			case "buttonColor":
				if err := json.Unmarshal(val, &v.ButtonColor); err != nil {
					return err
				}
			case "labelColor":
				if err := json.Unmarshal(val, &v.LabelColor); err != nil {
					return err
				}
			case "backgroundImage":
				if err := json.Unmarshal(val, &v.BackgroundImage); err != nil {
					return err
				}
			case "fonts":
				if err := json.Unmarshal(val, &v.Fonts); err != nil {
					return err
				}
			case "fontSizes":
				if err := json.Unmarshal(val, &v.FontSizes); err != nil {
					return err
				}
			}
		} else {
			var value interface{}
			if err := json.Unmarshal(val, &value); err != nil {
				return err
			}
			v.Custom[key] = value
			log.Printf("[DEBUG] Stored custom variable: %s = %v (type %T)", key, value, value)
		}
	}
	log.Printf("[DEBUG] Total custom variables: %+v", v.Custom)
	return nil
}

func (v *Variables) Get(name string) string {
	targetKey := strings.ToLower(name)
	log.Printf("[DEBUG] === Searching for variable: '%s' ===", name)

	// Check custom variables first
	for key, val := range v.Custom {
		if strings.EqualFold(key, targetKey) {
			log.Printf("[DEBUG] Found in custom vars: %s = %v (type %T)", key, val, val)
			switch val := val.(type) {
			case string:
				return val
			case float64:
				return strconv.FormatFloat(val, 'f', -1, 64)
			case int:
				return strconv.Itoa(val)
			default:
				return fmt.Sprintf("%v", val)
			}
		}
	}

	// Predefined variables
	switch targetKey {
	case "buttoncolor":
		log.Printf("[DEBUG] Found button color")
		return fmt.Sprintf("%d,%d,%d", v.ButtonColor.R, v.ButtonColor.G, v.ButtonColor.B)
	case "labelcolor":
		log.Printf("[DEBUG] Found label color")
		return fmt.Sprintf("%d,%d,%d", v.LabelColor.R, v.LabelColor.G, v.LabelColor.B)
	case "backgroundimage":
		log.Printf("[DEBUG] Found background image")
		return v.BackgroundImage
	}

	// Fonts
	for key, path := range v.Fonts {
		if strings.EqualFold(key, targetKey) {
			log.Printf("[DEBUG] Found font: %s", path)
			return path
		}
	}

	// Font sizes
	for key, size := range v.FontSizes {
		if strings.EqualFold(key, targetKey) {
			log.Printf("[DEBUG] Found font size: %d", size)
			return strconv.Itoa(size)
		}
	}

	log.Printf("[ERROR] MISSING VARIABLE: %s (searched as: %s)", name, targetKey)
	log.Printf("[DEBUG] Custom vars: %+v", v.Custom)
	log.Printf("[DEBUG] Fonts: %+v", v.Fonts)
	log.Printf("[DEBUG] Font sizes: %+v", v.FontSizes)
	return "MISSING_VAR"
}

func loadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	// Ensure fonts exist
	if config.Variables.Fonts == nil {
		config.Variables.Fonts = make(map[string]string)
	}
	if config.Variables.FontSizes == nil {
		config.Variables.FontSizes = make(map[string]int)
	}

	return &config, nil
}

func resolveColor(config *Config, colorName string, defaultColor sdl.Color) sdl.Color {
	if strings.HasPrefix(colorName, "$") {
		colorValue := config.Variables.Get(colorName[1:])
		parts := strings.Split(colorValue, ",")
		if len(parts) == 3 {
			r, _ := strconv.Atoi(parts[0])
			g, _ := strconv.Atoi(parts[1])
			b, _ := strconv.Atoi(parts[2])
			return sdl.Color{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
		}
		return defaultColor
	}

	if colorName != "" {
		r, g, b := hexToRGB(colorName)
		return sdl.Color{R: r, G: g, B: b, A: 255}
	}
	return defaultColor
}

func hexToRGB(hex string) (uint8, uint8, uint8) {
	if len(hex) == 7 {
		hex = hex[1:]
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return uint8(r), uint8(g), uint8(b)
}

func resolveBackground(renderer *sdl.Renderer, config *Config) *sdl.Texture {
	if config.Variables.BackgroundImage != "" {
		texture, err := img.LoadTexture(renderer, config.Variables.BackgroundImage)
		if err != nil {
			log.Printf("Failed to load background texture: %v", err)
			return nil
		}
		return texture
	}

	// Default background
	renderer.SetDrawColor(32, 32, 32, 255)
	renderer.Clear()
	return nil
}

func substituteVariables(text string, config *Config) string {
	return regexp.MustCompile(`\$(\w+)`).ReplaceAllStringFunc(text, func(m string) string {
		varName := m[1:]
		value := config.Variables.Get(varName)
		if value == "" {
			log.Printf("MISSING VARIABLE: %s", varName)
			return "MISSING_VAR"
		}
		return value
	})
}

func getFontAndSize(config *Config, fontName string) (*ttf.Font, int) {
	// Ensure maps are initialized
	if config.Variables.Fonts == nil {
		config.Variables.Fonts = make(map[string]string)
	}
	if config.Variables.FontSizes == nil {
		config.Variables.FontSizes = make(map[string]int)
	}

	// Get font path (case-insensitive)
	fontPath := "Roboto-Black.ttf" // Default fallback
	size := 24

	// Find matching font key
	for key, path := range config.Variables.Fonts {
		if strings.EqualFold(key, fontName) {
			fontPath = path
			break
		}
	}

	// Find matching font size key
	for key, val := range config.Variables.FontSizes {
		if strings.EqualFold(key, fontName) {
			size = val
			break
		}
	}

	font, err := ttf.OpenFont(fontPath, size)
	if err != nil {
		log.Printf("Error loading font %s: %v", fontPath, err)
		return nil, 0
	}
	return font, size
}

func handleInputElement(renderer *sdl.Renderer, config *Config, element Element) {
	virtualKeyboardActive = true
	defer func() { virtualKeyboardActive = false }()
	exitInput := false

	for !exitInput {
		renderer.SetDrawColor(249, 249, 249, 255)
		renderer.Clear()
		renderScene(renderer, config, config.Scenes[currentSceneIndex])
		renderKeyboard(renderer, config)
		renderer.Present()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					switch e.Keysym.Sym {
					case sdl.K_ESCAPE:
						exitInput = true
					case sdl.K_RETURN:
						handleKeyboardInput(config)
						exitInput = true
					}
				}
			}
		}
	}
}

func maxInt(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func getTextDimensions(font *ttf.Font, text string) (int32, int32) {
	if text == "" {
		return 0, 0
	}

	if font == nil {
		// Create a temporary font if none provided
		tempFont, err := ttf.OpenFont("Roboto-Black.ttf", 24)
		if err != nil {
			return 0, 0
		}
		defer tempFont.Close()
		font = tempFont
	}

	width, height, err := font.SizeUTF8(text)
	if err != nil {
		return 0, 0
	}
	return int32(width), int32(height)
}

func renderMenu(renderer *sdl.Renderer, config *Config, element Element) {
	textColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	highlightColor := sdl.Color{R: 0, G: 200, B: 255, A: 255}
	bgColor := sdl.Color{R: 32, G: 32, B: 32, A: 200}
	highlightBgColor := sdl.Color{R: 0, G: 150, B: 255, A: 255}

	font, _ := getFontAndSize(config, "medium")
	if font == nil {
		return
	}
	defer font.Close()

	// Background bar
	renderer.SetDrawColor(bgColor.R, bgColor.G, bgColor.B, bgColor.A)
	renderer.FillRect(&sdl.Rect{X: 0, Y: element.Y, W: 1280, H: 50})

	buttonX := int32(10)
	menuButtonRects = make(map[int]sdl.Rect) // Reset the menu button rects

	for i, scene := range config.Scenes {
		isSelected := currentSceneIndex == i

		btnColor := textColor
		rectColor := bgColor
		if isSelected {
			btnColor = highlightColor
			rectColor = highlightBgColor
		}

		label := scene.Name
		textWidth, textHeight := getTextDimensions(font, label)

		// Calculate button dimensions with padding
		padding := int32(20)
		width := textWidth + padding*2
		height := int32(40) // Fixed height for menu buttons

		// Draw rounded rectangle as button background
		drawRoundedRect(renderer, &sdl.Rect{
			X: buttonX,
			Y: element.Y + 5,
			W: width,
			H: height,
		}, 10, rectColor)

		// Draw label centered in the button
		renderText(renderer, config, font, label, btnColor,
			buttonX+(width-textWidth)/2,
			element.Y+5+(height-textHeight)/2)

		menuButtonRects[i] = sdl.Rect{
			X: buttonX,
			Y: element.Y + 5,
			W: width,
			H: height,
		}
		buttonX += width + 10
	}

	// Clock display (right side)
	currentTime := time.Now().Format("15:04")
	renderText(renderer, config, font, currentTime, textColor, 1200, element.Y+15)
}

func renderText(renderer *sdl.Renderer, config *Config, font *ttf.Font, text string, color sdl.Color, x int32, y int32) (int32, int32) {
	processedText := substituteVariables(text, config)
	if processedText == "" || font == nil {
		return 0, 0
	}

	surface, err := font.RenderUTF8Blended(processedText, color)
	if err != nil {
		log.Printf("Render error: %v", err)
		return 0, 0
	}
	defer surface.Free()

	texture, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		log.Printf("Texture error: %v", err)
		return 0, 0
	}
	defer texture.Destroy()

	// Get exact dimensions from texture
	_, _, w, h, _ := texture.Query()
	renderer.Copy(texture, nil, &sdl.Rect{
		X: x,
		Y: y,
		W: w,
		H: h,
	})

	return w, h
}
func renderButton(renderer *sdl.Renderer, config *Config, element Element) {
	defaultTextColor := sdl.Color{R: 0, G: 0, B: 0, A: 255}     // Default to black
	defaultBgColor := sdl.Color{R: 255, G: 255, B: 255, A: 255} // Default to white

	color := resolveColor(config, element.Color, defaultTextColor)
	bgColor := resolveColor(config, element.BgColor, defaultBgColor)

	// Render button background
	renderer.SetDrawColor(bgColor.R, bgColor.G, bgColor.B, bgColor.A)
	renderer.FillRect(&sdl.Rect{X: element.X, Y: element.Y, W: 200, H: 50})

	// Render button text
	font, _ := getFontAndSize(config, element.Font)
	renderText(renderer, config, font, element.Text, color, element.X+100, element.Y+25)
}

func renderKeyboard(renderer *sdl.Renderer, config *Config) {
	if !virtualKeyboardActive {
		return
	}

	// Dark overlay
	renderer.SetDrawColor(0, 0, 0, 200)
	renderer.FillRect(&sdl.Rect{X: 0, Y: 0, W: 1280, H: 720})

	keyWidth := int32(60)
	keyHeight := int32(60)
	padding := int32(10)
	startX := (1280 - (10*keyWidth + 9*padding)) / 2
	startY := int32(200)

	for y, row := range keyboard {
		rowStartX := startX
		if y == 1 {
			rowStartX += keyWidth / 2
		}
		if y == 2 {
			rowStartX += keyWidth
		}
		if y == 3 {
			rowStartX += keyWidth * 3
		}

		for x, key := range row {
			// Draw key background
			bgColor := sdl.Color{R: 255, G: 255, B: 255}
			if x == keyboardPosX && y == keyboardPosY {
				bgColor = sdl.Color{R: 0, G: 255, B: 0}
			}
			renderer.SetDrawColor(bgColor.R, bgColor.G, bgColor.B, 255)

			rect := &sdl.Rect{
				X: rowStartX + int32(x)*(keyWidth+padding),
				Y: startY + int32(y)*(keyHeight+padding),
				W: keyWidth,
				H: keyHeight,
			}
			renderer.FillRect(rect)

			// Draw key text
			font, _ := getFontAndSize(config, "medium") // Now has access to config
			renderText(renderer, config, font, key, sdl.Color{R: 0, G: 0, B: 0},
				rect.X+keyWidth/2,
				rect.Y+keyHeight/2,
			)
		}
	}
}

func isInputButtonSelected(element Element) bool {
	return virtualKeyboardActive && inputActiveElement == &element
}

func handleTextInput(event *sdl.KeyboardEvent, config *Config) {
	switch event.Keysym.Sym {
	case sdl.K_BACKSPACE:
		if len(inputTextBuffer) > 0 {
			inputTextBuffer = inputTextBuffer[:len(inputTextBuffer)-1]
		}
	case sdl.K_RETURN, sdl.K_KP_ENTER:
		// Finish input
		inputActiveElement = nil
		sdl.StopTextInput()
	default:
		// Characters handled by TextInputEvent
	}
	updateInputVariable(config)
}

func handleInputSelection(renderer *sdl.Renderer, config *Config, element *Element) {
	inputActiveElement = element
	virtualKeyboardActive = true
	handleInputElement(renderer, config, *element)
	sdl.StartTextInput()
}

func updateInputVariable(config *Config) {
	if inputActiveElement != nil && inputActiveElement.Variable != "" {
		config.Variables.Custom[inputActiveElement.Variable] = inputTextBuffer
	}
}

func initKeyboard() {
	keyboard = [][]string{
		{"Q", "W", "E", "R", "T", "Y", "U", "I", "O", "P"},
		{"A", "S", "D", "F", "G", "H", "J", "K", "L"},
		{"Z", "X", "C", "V", "B", "N", "M"},
		{"SPACE", "BACK", "ENTER"},
	}
	keyboardPosX, keyboardPosY = 0, 0
}

func handleVirtualKeyboardInput(event *sdl.KeyboardEvent, config *Config) {
	switch event.Keysym.Sym {
	case sdl.K_UP:
		if keyboardPosY > 0 {
			keyboardPosY--
		}
	case sdl.K_DOWN:
		if keyboardPosY < len(keyboard)-1 {
			keyboardPosY++
		}
	case sdl.K_LEFT:
		if keyboardPosX > 0 {
			keyboardPosX--
		}
	case sdl.K_RIGHT:
		if keyboardPosX < len(keyboard[keyboardPosY])-1 {
			keyboardPosX++
		}
	case sdl.K_RETURN:
		handleKeyboardInput(config)
	}
}

func handleKeyboardInput(config *Config) {
	selectedKey := keyboard[keyboardPosY][keyboardPosX]
	switch selectedKey {
	case "SPACE":
		inputTextBuffer += " "
	case "BACK":
		if len(inputTextBuffer) > 0 {
			inputTextBuffer = inputTextBuffer[:len(inputTextBuffer)-1]
		}
	case "ENTER":
		virtualKeyboardActive = false
		inputActiveElement = nil
	default:
		inputTextBuffer += selectedKey
	}
	updateInputVariable(config)
}

func renderScene(renderer *sdl.Renderer, config *Config, sceneConfig SceneConfig) {
	log.Printf("Rendering scene: %s", sceneConfig.Name)
	fontCache := make(map[string]*ttf.Font)
	bgTexture := resolveBackground(renderer, config)
	if bgTexture != nil {
		renderer.Copy(bgTexture, nil, &sdl.Rect{X: 0, Y: 0, W: 1280, H: 720})
		bgTexture.Destroy()
	}

	for i, element := range sceneConfig.Elements {
		log.Printf("Rendering element %d: %s (%s)", i, element.Text, element.Type)
		defaultTextColor := sdl.Color{R: 0, G: 0, B: 0, A: 255}     // Default to black
		defaultBgColor := sdl.Color{R: 255, G: 255, B: 255, A: 255} // Default to white

		color := resolveColor(config, element.Color, defaultTextColor)
		bgColor := resolveColor(config, element.BgColor, defaultBgColor)

		// Highlight selected button by inverting colors
		if element.Type == "button" && i == selectedButtonIndex {
			color, bgColor = bgColor, color // Invert colors
		}

		font, _ := fontCache[element.Font]
		if font == nil {
			font, _ = getFontAndSize(config, element.Font)
			fontCache[element.Font] = font
		}

		switch element.Type {
		case "image":
			log.Printf("Loading image: %s", element.Image)
			if element.Image != "" {
				// Handle width with variable substitution
				widthStr := substituteVariables(string(element.Width), config)
				width, err := strconv.Atoi(widthStr)
				if err != nil {
					width = 0
				}

				// Handle height with variable substitution
				heightStr := substituteVariables(string(element.Height), config)
				height, err := strconv.Atoi(heightStr)
				if err != nil {
					height = 0
				}

				imageTexture, err := img.LoadTexture(renderer, element.Image)
				if err == nil {
					defer imageTexture.Destroy()
					imageRect := sdl.Rect{X: element.X, Y: element.Y, W: int32(width), H: int32(height)}
					renderer.Copy(imageTexture, nil, &imageRect)
				}
			}
		case "input":
			log.Printf("Rendering input field at (%d,%d)", element.X, element.Y)
			renderInputField(renderer, config, element)
		case "video":
			if !videoPlayed && element.Video != "" {
				widthStr := substituteVariables(string(element.Width), config)
				width, err := strconv.Atoi(widthStr)
				if err != nil {
					width = 0
				}

				heightStr := substituteVariables(string(element.Height), config)
				height, err := strconv.Atoi(heightStr)
				if err != nil {
					height = 0
				}

				// Construct the ffplay command with the correct parameters
				cmd := exec.Command("ffmpeg/ffplay", element.Video,
					"-noborder",
					"-x", strconv.Itoa(width),
					"-y", strconv.Itoa(height),
					"-left", strconv.Itoa(int(element.X)),
					"-top", strconv.Itoa(int(element.Y)),
					"-autoexit")

				// Start the ffplay process
				if err := cmd.Start(); err != nil {
					fmt.Println("Error starting ffplay:", err)
				}
				videoPlayed = true

				/*if err := cmd.Wait(); err != nil {
					fmt.Println("Error waiting for ffplay:", err)
				}*/
			}
		case "label": // Add specific label handling
			if font != nil {
				renderText(renderer, config, font, element.Text, color,
					element.X,
					element.Y)
			}
		case "collapsedlist":
			if element.Command != "" && element.ListVariable != "" {
				// Execute command and parse output if not already done
				if _, exists := config.Variables.Custom[element.ListVariable]; !exists {
					executeCommandAndParse(config, element.Command, element.ListVariable)
				}

				// Render the collapsed list
				if items, ok := config.Variables.Custom[element.ListVariable].([]CollapsedListItem); ok {
					renderCollapsedList(renderer, config, element, items)
				}
			} else {
				// Render placeholder if no command is set
				renderText(renderer, config, font, "Collapsed List", color, element.X, element.Y)
			}
		// In the renderScene function, update the button rendering case:
		case "button":
			// Calculate dimensions
			textWidth, textHeight := getTextDimensions(font, element.Text)
			width := textWidth + 20
			height := textHeight + 10

			// Override dimensions if specified
			if string(element.Width) != "" {
				widthStr := substituteVariables(string(element.Width), config)
				w, _ := strconv.Atoi(widthStr)
				width = int32(w)
			}
			if string(element.Height) != "" {
				heightStr := substituteVariables(string(element.Height), config)
				h, _ := strconv.Atoi(heightStr)
				height = int32(h)
			}

			// Render button background
			renderer.SetDrawColor(bgColor.R, bgColor.G, bgColor.B, bgColor.A)
			renderer.FillRect(&sdl.Rect{X: element.X, Y: element.Y, W: width, H: height})

			// Render button text - centered properly
			textX := element.X + (width-textWidth)/2
			textY := element.Y + (height-textHeight)/2
			renderText(renderer, config, font, element.Text, color, textX, textY)
		case "menu":
			renderMenu(renderer, config, element)
		default:
			log.Printf("Unknown element type: %s", element.Type)
		}
	}

	for _, font := range fontCache {
		font.Close()
	}
}

func renderCollapsedList(renderer *sdl.Renderer, config *Config, element Element, items []CollapsedListItem) {
	// Render the list icon
	font, _ := getFontAndSize(config, element.Font)
	if font == nil {
		return
	}
	defer font.Close()

	// Draw list background
	renderer.SetDrawColor(240, 240, 240, 255)
	renderer.FillRect(&sdl.Rect{X: element.X, Y: element.Y, W: 300, H: 40})

	// Draw list icon and text
	renderText(renderer, config, font, "📋 Collapsed List",
		resolveColor(config, element.Color, sdl.Color{R: 0, G: 0, B: 0, A: 255}),
		element.X+10, element.Y+20)

	for i, item := range items {
		yPos := element.Y + 40 + int32(i)*60
		renderCollapsedListItem(renderer, config, item, element.X, yPos, 300, 60)
	}

}

func drawRoundedRect(renderer *sdl.Renderer, rect *sdl.Rect, radius int32, color sdl.Color) {
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)

	// Draw the main rectangle (excluding corners)
	renderer.FillRect(&sdl.Rect{
		X: rect.X + radius,
		Y: rect.Y,
		W: rect.W - 2*radius,
		H: rect.H,
	})
	renderer.FillRect(&sdl.Rect{
		X: rect.X,
		Y: rect.Y + radius,
		W: rect.W,
		H: rect.H - 2*radius,
	})

	// Draw the rounded corners
	drawFilledCircle(renderer, rect.X+radius, rect.Y+radius, radius, color)
	drawFilledCircle(renderer, rect.X+rect.W-radius, rect.Y+radius, radius, color)
	drawFilledCircle(renderer, rect.X+radius, rect.Y+rect.H-radius, radius, color)
	drawFilledCircle(renderer, rect.X+rect.W-radius, rect.Y+rect.H-radius, radius, color)
}

func renderCollapsedListItem(renderer *sdl.Renderer, config *Config, item CollapsedListItem, x, y, width, height int32) {
	// Draw item background
	renderer.SetDrawColor(255, 255, 255, 255)
	renderer.FillRect(&sdl.Rect{X: x, Y: y, W: width, H: height})

	// Draw item image if available
	if item.Image != "" {
		texture, err := img.LoadTexture(renderer, item.Image)
		if err == nil {
			defer texture.Destroy()
			renderer.Copy(texture, nil, &sdl.Rect{X: x + 5, Y: y + 5, W: 50, H: 50})
		}
	}

	// Draw item text
	font, _ := getFontAndSize(config, "small")
	if font != nil {
		defer font.Close()
		renderText(renderer, config, font, item.Title,
			sdl.Color{R: 0, G: 0, B: 0, A: 255},
			x+60, y+10)

		if item.Description != "" {
			renderText(renderer, config, font, item.Description,
				sdl.Color{R: 100, G: 100, B: 100, A: 255},
				x+60, y+30)
		}
	}
}

func renderInputField(renderer *sdl.Renderer, config *Config, element Element) {
	// Draw background
	bgColor := resolveColor(config, element.BgColor, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	renderer.SetDrawColor(bgColor.R, bgColor.G, bgColor.B, bgColor.A)
	renderer.FillRect(&sdl.Rect{
		X: element.X,
		Y: element.Y,
		W: 300, // Default width
		H: 40,  // Default height
	})

	// Draw text
	textColor := resolveColor(config, element.Color, sdl.Color{R: 0, G: 0, B: 0, A: 255})
	font, _ := getFontAndSize(config, element.Font)

	// Show cursor if active
	text := inputTextBuffer
	if inputActiveElement == &element && (uint32(sdl.GetTicks64()/500)%2 == 0) {
		text += "_"
	}

	renderText(renderer, config, font, text, textColor, element.X+10, element.Y+20)

	// Draw border if active
	if inputActiveElement == &element {
		renderer.SetDrawColor(0, 120, 215, 255)
		renderer.DrawRect(&sdl.Rect{
			X: element.X - 2,
			Y: element.Y - 2,
			W: 304,
			H: 44,
		})
	}
}

func renderImage(renderer *sdl.Renderer, config *Config, element Element) {
	if element.Image == "" {
		return
	}
	texture, err := img.LoadTexture(renderer, element.Image)
	if err != nil {
		log.Printf("Failed to load image %s: %v", element.Image, err)
		return
	}
	defer texture.Destroy()

	widthStr := substituteVariables(string(element.Width), config)
	width, _ := strconv.Atoi(widthStr)
	heightStr := substituteVariables(string(element.Height), config)
	height, _ := strconv.Atoi(heightStr)

	if width == 0 || height == 0 {
		_, _, w, h, _ := texture.Query()
		width = int(w)
		height = int(h)
	}
	renderer.Copy(texture, nil, &sdl.Rect{X: element.X, Y: element.Y, W: int32(width), H: int32(height)})
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Unhandled error: %v\n", r)
			log.Println("Stack trace:")
			debug.PrintStack()
			os.Exit(-1)
		}
	}()

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_JOYSTICK | sdl.INIT_GAMECONTROLLER); err != nil {
		fmt.Println("Error initializing SDL:", err)
		os.Exit(1)
	}
	defer sdl.Quit()

	if err := ttf.Init(); err != nil {
		fmt.Println("Error initializing TTF:", err)
		os.Exit(1)
	}
	defer ttf.Quit()

	if err := img.Init(img.INIT_PNG | img.INIT_JPG | img.INIT_TIF); err != nil {
		fmt.Println("Error initializing IMG:", err)
		os.Exit(1)
	}
	defer img.Quit()

	config, err := loadConfig("jukaconfig.json")
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	// Auto-select the first selectable element in the initial scene
	firstSelectable := findFirstSelectableElement(config.Scenes[currentSceneIndex])
	if firstSelectable != -1 {
		selectedButtonIndex = firstSelectable
	}

	screenWidth := int32(1280)
	screenHeight := int32(720)

	window, err := sdl.CreateWindow(config.Title, sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, screenWidth, screenHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println("Error creating window:", err)
		os.Exit(1)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println("Error creating renderer:", err)
		os.Exit(1)
	}
	defer renderer.Destroy()

	mapping1 := "030000005e0400008e02000014010000,X360 Controller,a:b0,b:b1,back:b6,dpdown:h0.4,dpleft:h0.8,dpright:h0.2,dpup:h0.1,guide:b8,leftshoulder:b4,leftstick:b9,lefttrigger:a2,leftx:a0,lefty:a1,rightshoulder:b5,rightstick:b10,righttrigger:a5,rightx:a3,righty:a4,start:b7,x:b2,y:b3,platform:Linux,"
	mapping2 := "0000000058626f782047616d65706100,Xbox Gamepad (userspace driver),platform:Linux,a:b0,b:b1,x:b2,y:b3,start:b7,back:b6,guide:b8,dpup:h0.1,dpdown:h0.4,dpleft:h0.8,dpright:h0.2,leftshoulder:b4,rightshoulder:b5,lefttrigger:a5,righttrigger:a4,leftstick:b9,rightstick:b10,leftx:a0,lefty:a1,rightx:a2,righty:a3,"

	if sdl.GameControllerAddMapping(mapping1) == -1 {
		fmt.Printf("Failed to add controller mapping: %s\n", sdl.GetError())
	}

	if sdl.GameControllerAddMapping(mapping2) == -1 {
		fmt.Printf("Failed to add controller mapping: %s\n", sdl.GetError())
	}

	if sdl.NumJoysticks() > 0 {
		if controller := sdl.GameControllerOpen(0); controller != nil {
			defer controller.Close()
			fmt.Println("Controller detected.")
		}
	}

	initKeyboard() // Initialize virtual keyboard layout

	currentSceneIndex = 0
	selectedButtonIndex = 0
	var inputText = ""

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.KeyboardEvent: // Use pointer receiver
				if e.Type == sdl.KEYDOWN {
					if virtualKeyboardActive {
						handleVirtualKeyboardInput(e, config)
					} else if inputActiveElement != nil {
						// Handle direct text input
						handleTextInput(e, config)
					} else {
						// Handle menu and other navigation
						switch e.Keysym.Sym {
						case sdl.K_UP:
							moveSelection(config, -1)
						case sdl.K_DOWN:
							moveSelection(config, 1)
						case sdl.K_LEFT:
							moveSelection(config, -1)
						case sdl.K_RIGHT:
							moveSelection(config, 1)
						case sdl.K_RETURN, sdl.K_SPACE:
							if selectedButtonIndex >= 0 && selectedButtonIndex < len(config.Scenes[currentSceneIndex].Elements) {
								selectedElement := config.Scenes[currentSceneIndex].Elements[selectedButtonIndex]
								if selectedElement.Type == "input" {
									handleInputSelection(renderer, config, &selectedElement)
								} else if selectedElement.Type == "menu" {
									// Menu is already handled by left/right navigation
								} else {
									triggerSelectedElement(renderer, config)
								}
							}
						// Add Q/E for menu navigation
						case sdl.K_q:
							changeScene(config, -1)
						case sdl.K_e:
							changeScene(config, 1)
						}
					}
				}
			case *sdl.TextInputEvent: // Use pointer receiver
				inputText += string(e.Text[:])
				updateInputVariable(config)
			case *sdl.QuitEvent: // Use pointer receiver
				running = false
			case *sdl.MouseButtonEvent:
				if e.Button == sdl.BUTTON_LEFT && e.Type == sdl.MOUSEBUTTONDOWN {
					mouseX, mouseY := int32(e.X), int32(e.Y)

					// Check menu buttons first
					for sceneIndex, rect := range menuButtonRects {
						if mouseX >= rect.X && mouseX <= rect.X+rect.W &&
							mouseY >= rect.Y && mouseY <= rect.Y+rect.H {
							// Change scene on click
							currentSceneIndex = sceneIndex

							// Auto-select the first selectable element in the new scene
							firstSelectable := findFirstSelectableElement(config.Scenes[currentSceneIndex])
							if firstSelectable != -1 {
								selectedButtonIndex = firstSelectable
							} else {
								selectedButtonIndex = 0
							}

							videoPlayed = false
							break // Exit after handling the click
						}
					}

					// If not a menu button, check other elements
					currentScene := config.Scenes[currentSceneIndex]
					for i, element := range currentScene.Elements {
						// Input field handling
						if element.Type == "input" {
							widthStr := substituteVariables(string(element.Width), config)
							width, _ := strconv.Atoi(widthStr)
							if width == 0 {
								width = 200
							}
							heightStr := substituteVariables(string(element.Height), config)
							height, _ := strconv.Atoi(heightStr)
							if height == 0 {
								height = 50
							}

							if mouseX >= element.X && mouseX <= element.X+int32(width) &&
								mouseY >= element.Y && mouseY <= element.Y+int32(height) {
								handleInputSelection(renderer, config, &currentScene.Elements[i])
							}
						} else if element.Type == "button" {
							textWidth, textHeight := getTextDimensions(nil, element.Text)
							btnWidth := textWidth + 20
							btnHeight := textHeight + 10

							if string(element.Width) != "" {
								widthStr := substituteVariables(string(element.Width), config)
								w, _ := strconv.Atoi(widthStr)
								btnWidth = int32(w)
							}
							if string(element.Height) != "" {
								heightStr := substituteVariables(string(element.Height), config)
								h, _ := strconv.Atoi(heightStr)
								btnHeight = int32(h)
							}

							if mouseX >= element.X && mouseX <= element.X+btnWidth &&
								mouseY >= element.Y && mouseY <= element.Y+btnHeight {
								handleTrigger(renderer, config, element)
							}
						}
					}
				}

			case *sdl.ControllerButtonEvent: // Use pointer receiver
				if e.Type == sdl.CONTROLLERBUTTONDOWN {
					if virtualKeyboardActive {
						switch e.Button {
						case sdl.CONTROLLER_BUTTON_DPAD_UP:
							if keyboardPosY > 0 {
								keyboardPosY--
							}
						case sdl.CONTROLLER_BUTTON_DPAD_DOWN:
							if keyboardPosY < len(keyboard)-1 {
								keyboardPosY++
							}
						case sdl.CONTROLLER_BUTTON_DPAD_LEFT:
							if keyboardPosX > 0 {
								keyboardPosX--
							}
						case sdl.CONTROLLER_BUTTON_DPAD_RIGHT:
							if keyboardPosX < len(keyboard[keyboardPosY])-1 {
								keyboardPosX++
							}
						case sdl.CONTROLLER_BUTTON_A, sdl.CONTROLLER_BUTTON_B:
							handleKeyboardInput(config)
						}
					} else {
						switch e.Button {
						case sdl.CONTROLLER_BUTTON_DPAD_UP:
							moveSelection(config, -1)
						case sdl.CONTROLLER_BUTTON_DPAD_DOWN:
							moveSelection(config, 1)
						case sdl.CONTROLLER_BUTTON_DPAD_LEFT:
							moveSelection(config, -1)
						case sdl.CONTROLLER_BUTTON_DPAD_RIGHT:
							moveSelection(config, 1)
						case sdl.CONTROLLER_BUTTON_A, sdl.CONTROLLER_BUTTON_B:
							if selectedButtonIndex >= 0 && selectedButtonIndex < len(config.Scenes[currentSceneIndex].Elements) {
								selectedElement := config.Scenes[currentSceneIndex].Elements[selectedButtonIndex]
								if selectedElement.Type == "input" {
									handleInputSelection(renderer, config, &selectedElement)
								} else if selectedElement.Type == "menu" {
									// Menu is already handled by left/right navigation
								} else {
									triggerSelectedElement(renderer, config)
								}
							}
						// Add shoulder buttons for menu navigation
						case sdl.CONTROLLER_BUTTON_LEFTSHOULDER:
							changeScene(config, -1)
						case sdl.CONTROLLER_BUTTON_RIGHTSHOULDER:
							changeScene(config, 1)
						}
					}
				}
				//default:
				//fmt.Printf("Unhandled event: %T\n", event)
			}
		}

		renderScene(renderer, config, config.Scenes[currentSceneIndex])
		renderer.Present()
	}
}

func changeScene(config *Config, direction int) {
	currentSceneIndex += direction
	if currentSceneIndex < 0 {
		currentSceneIndex = len(config.Scenes) - 1
	} else if currentSceneIndex >= len(config.Scenes) {
		currentSceneIndex = 0
	}

	// Auto-select the first selectable element in the new scene
	firstSelectable := findFirstSelectableElement(config.Scenes[currentSceneIndex])
	if firstSelectable != -1 {
		selectedButtonIndex = firstSelectable
	} else {
		selectedButtonIndex = 0
	}

	videoPlayed = false
}

func moveSelection(config *Config, direction int) {
	currentScene := config.Scenes[currentSceneIndex]
	elements := currentScene.Elements

	// Create a list of navigable element indices (only buttons and inputs, skip menus)
	var interactive []int
	for i, el := range elements {
		if (el.Type == "button" || el.Type == "input") && el.Type != "menu" {
			interactive = append(interactive, i)
		}
	}

	// No interactive elements available
	if len(interactive) == 0 {
		selectedButtonIndex = -1
		return
	}

	// Find current position in interactive list
	currentIdx := -1
	for idx, val := range interactive {
		if val == selectedButtonIndex {
			currentIdx = idx
			break
		}
	}

	// Handle new selections or wrap around
	if currentIdx == -1 {
		selectedButtonIndex = interactive[0]
		return
	}

	// Calculate new index with wrap-around
	newIdx := currentIdx + direction
	if newIdx >= len(interactive) {
		newIdx = 0
	} else if newIdx < 0 {
		newIdx = len(interactive) - 1
	}

	// Update selection
	selectedButtonIndex = interactive[newIdx]
}

func triggerSelectedElement(renderer *sdl.Renderer, config *Config) {
	selectedElement := config.Scenes[currentSceneIndex].Elements[selectedButtonIndex]
	if selectedElement.Type == "button" {
		handleTrigger(renderer, config, selectedElement) // Pass renderer here
	}
}

func handleTrigger(renderer *sdl.Renderer, config *Config, element Element) {
	if element.Trigger == "" {
		return
	}

	switch element.Trigger {
	case "set_variable":
		if element.TriggerTarget != "" {
			config.Variables.Custom[element.TriggerTarget] = element.TriggerValue
		}

	case "external_app":
		cmd := exec.Command(element.TriggerTarget)
		if err := cmd.Start(); err != nil {
			log.Printf("Failed to start external app: %v", err)
		}

	case "play_video":
		go func() {
			cmd := exec.Command("ffplay", element.TriggerTarget, "-fs", "-autoexit")
			if err := cmd.Run(); err != nil {
				log.Printf("Video playback failed: %v", err)
			}
		}()

	case "play_image":
		texture, err := img.LoadTexture(renderer, element.TriggerTarget) // Use passed renderer
		if err == nil {
			renderer.Copy(texture, nil, nil)
			renderer.Present()
		}
		defer texture.Destroy()
		renderer.Copy(texture, nil, nil)
		renderer.Present()
		sdl.Delay(3000)
	case "exit":
		//fmt.Println("Exiting the application.")
		os.Exit(0)
	case "change_scene":
		if element.TriggerTarget != "" {
			//fmt.Println("Changing scene to:", element.TriggerTarget)
			for i, scene := range config.Scenes {
				if scene.Name == element.TriggerTarget {
					currentSceneIndex = i
					selectedButtonIndex = 0
					videoPlayed = false
					//fmt.Println("Scene changed to:", element.TriggerTarget)
					break
				}
			}
		}
	}
}

func drawFilledCircle(renderer *sdl.Renderer, x, y, r int32, color sdl.Color) {
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	for w := int32(0); w < r*2; w++ {
		for h := int32(0); h < r*2; h++ {
			dx := r - w
			dy := r - h
			if dx*dx+dy*dy <= r*r {
				renderer.DrawPoint(x+dx, y+dy)
			}
		}
	}
}

func setConfigVariable(config *Config, variableName, value string) {
	//fmt.Println("Setting configuration variable:", variableName, "to", value)
	switch variableName {
	case "buttonColor":
		colorParts := strings.Split(value, ",")
		if len(colorParts) == 3 {
			r, _ := strconv.Atoi(colorParts[0])
			g, _ := strconv.Atoi(colorParts[1])
			b, _ := strconv.Atoi(colorParts[2])
			config.Variables.ButtonColor = struct {
				R int `json:"r"`
				G int `json:"g"`
				B int `json:"b"`
			}{r, g, b}
		}
	case "labelColor":
		colorParts := strings.Split(value, ",")
		if len(colorParts) == 3 {
			r, _ := strconv.Atoi(colorParts[0])
			g, _ := strconv.Atoi(colorParts[1])
			b, _ := strconv.Atoi(colorParts[2])
			config.Variables.LabelColor = struct {
				R int `json:"r"`
				G int `json:"g"`
				B int `json:"b"`
			}{r, g, b}
		}
	case "backgroundImage":
		config.Variables.BackgroundImage = value
	}
}

func findFirstSelectableElement(scene SceneConfig) int {
	for i, element := range scene.Elements {
		if element.Type == "button" || element.Type == "input" {
			return i
		}
	}
	return -1
}
