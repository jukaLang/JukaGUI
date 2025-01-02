package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

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
	BackgroundImage string `json:"backgroundImage"`
	Fonts           struct {
		Title  string `json:"title"`
		Big    string `json:"big"`
		Medium string `json:"medium"`
		Small  string `json:"small"`
	} `json:"fonts"`
	FontSizes struct {
		Title  int `json:"title"`
		Big    int `json:"big"`
		Medium int `json:"medium"`
		Small  int `json:"small"`
	} `json:"fontSizes"`
}

type SceneConfig struct {
	Name       string    `json:"name"`
	Background any       `json:"background"`
	Elements   []Element `json:"elements"`
}

type Element struct {
	Type          string `json:"type"`
	Text          string `json:"text"`
	Color         string `json:"color"`
	X             int32  `json:"x"`
	Y             int32  `json:"y"`
	Font          string `json:"font"`
	BgColor       string `json:"bgColor"`
	Trigger       string `json:"trigger"`
	TriggerTarget string `json:"triggerTarget"`
	TriggerValue  string `json:"triggerValue"`
	Image         string `json:"image"`
	Width         string `json:"width"`
	Height        string `json:"height"`
}

var currentSceneIndex int
var selectedButtonIndex int

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

	return &config, nil
}

func resolveColor(config *Config, colorName string, defaultColor sdl.Color) sdl.Color {
	if strings.HasPrefix(colorName, "$") {
		color := config.Variables.ButtonColor
		if colorName[1:] == "labelColor" {
			color = config.Variables.LabelColor
		}
		return sdl.Color{R: uint8(color.R), G: uint8(color.G), B: uint8(color.B), A: 255}
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

func resolveBackground(renderer *sdl.Renderer, config *Config, background any) *sdl.Texture {
	if str, ok := background.(string); ok {
		if strings.HasPrefix(str, "$") {
			texturePath := config.Variables.BackgroundImage
			texture, err := img.LoadTexture(renderer, texturePath)
			if err != nil {
				os.Exit(1)
			}
			return texture
		}
	} else if bgMap, ok := background.(map[string]any); ok {
		renderer.SetDrawColor(uint8(bgMap["r"].(float64)), uint8(bgMap["g"].(float64)), uint8(bgMap["b"].(float64)), 255)
		renderer.Clear()
		return nil
	}
	renderer.SetDrawColor(249, 249, 249, 255) // Default to #f9f9f9
	renderer.Clear()
	return nil
}

func getFontAndSize(config *Config, fontName string) (*ttf.Font, int) {
	var fontPath string
	var size int
	switch fontName {
	case "title":
		fontPath = config.Variables.Fonts.Title
		size = config.Variables.FontSizes.Title
	case "big":
		fontPath = config.Variables.Fonts.Big
		size = config.Variables.FontSizes.Big
	case "medium":
		fontPath = config.Variables.Fonts.Medium
		size = config.Variables.FontSizes.Medium
	case "small":
		fontPath = config.Variables.Fonts.Small
		size = config.Variables.FontSizes.Small
	default:
		fontPath = config.Variables.Fonts.Medium // Default font
		size = 24                                // Default font size
	}
	font, err := ttf.OpenFont(fontPath, size)
	if err != nil {
		os.Exit(1)
	}
	return font, size
}

func renderText(renderer *sdl.Renderer, font *ttf.Font, text string, color sdl.Color, x int32, y int32) (int32, int32) {
	if text == "" || font == nil {
		return 0, 0
	}

	// Render text to surface
	surface, err := font.RenderUTF8Blended(text, color)
	if err != nil {
		os.Exit(1)
	}
	defer surface.Free()

	if surface.W == 0 && surface.H == 0 {
		return 0, 0
	}

	// Create texture from surface
	texture, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		os.Exit(1)
	}
	defer texture.Destroy()

	// Render texture to screen using surface dimensions
	textWidth := int32(surface.W)
	textHeight := int32(surface.H)
	renderer.Copy(texture, nil, &sdl.Rect{X: x - textWidth/2, Y: y - textHeight/2, W: textWidth, H: textHeight})

	return textWidth, textHeight
}

func renderScene(renderer *sdl.Renderer, config *Config, sceneConfig SceneConfig) {
	fontCache := make(map[string]*ttf.Font)
	bgTexture := resolveBackground(renderer, config, sceneConfig.Background)
	if bgTexture != nil {
		renderer.Copy(bgTexture, nil, &sdl.Rect{X: 0, Y: 0, W: 1280, H: 720})
		bgTexture.Destroy()
	}

	for i, element := range sceneConfig.Elements {
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

		if element.Type == "image" && element.Image != "" {
			imageTexture, err := img.LoadTexture(renderer, element.Image)
			if err == nil {
				defer imageTexture.Destroy()
				width, err := strconv.Atoi(element.Width)
				if err != nil {
					width = 0 // Default width if conversion fails
				}
				height, err := strconv.Atoi(element.Height)
				if err != nil {
					height = 0 // Default height if conversion fails
				}
				imageRect := sdl.Rect{X: element.X, Y: element.Y, W: int32(width), H: int32(height)}
				renderer.Copy(imageTexture, nil, &imageRect)
			}
		} else {
			// Calculate text dimensions
			textWidth, textHeight := getTextDimensions(font, element.Text)
			width := textWidth + 20   // Add padding to text width
			height := textHeight + 10 // Add padding to text height

			// Override width and height if specified
			if element.Width != "" {
				w, err := strconv.Atoi(element.Width)
				if err == nil {
					width = int32(w)
				}
			}
			if element.Height != "" {
				h, err := strconv.Atoi(element.Height)
				if err == nil {
					height = int32(h)
				}
			}

			// Render button background or label background
			renderer.SetDrawColor(bgColor.R, bgColor.G, bgColor.B, bgColor.A)
			renderer.FillRect(&sdl.Rect{X: element.X, Y: element.Y, W: width, H: height})

			// Render button text or label text
			renderText(renderer, font, element.Text, color, element.X+width/2, element.Y+height/2)
		}
	}

	for _, font := range fontCache {
		font.Close()
	}
}

// Helper function to get text dimensions
func getTextDimensions(font *ttf.Font, text string) (int32, int32) {
	surface, err := font.RenderUTF8Blended(text, sdl.Color{R: 0, G: 0, B: 0, A: 0})
	if err != nil {
		fmt.Println("Error rendering text to surface:", err)
		os.Exit(1)
	}
	defer surface.Free()

	return int32(surface.W), int32(surface.H)
}

func main() {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_GAMECONTROLLER | sdl.INIT_HAPTIC | sdl.INIT_JOYSTICK | sdl.INIT_AUDIO); err != nil {
		fmt.Println("Error initializing SDL:", err)
		os.Exit(1)
	}
	defer sdl.Quit()

	if err := ttf.Init(); err != nil {
		fmt.Println("Error initializing TTF:", err)
		os.Exit(1)
	}
	defer ttf.Quit()

	if err := img.Init(img.INIT_PNG); err != nil {
		fmt.Println("Error initializing IMG:", err)
		os.Exit(1)
	}
	defer img.Quit()

	config, err := loadConfig("jukaconfig.json")
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
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
		fmt.Println("Failed to add controller mapping: %s\n", sdl.GetError())
	}

	if sdl.GameControllerAddMapping(mapping2) == -1 {
		fmt.Println("Failed to add controller mapping: %s\n", sdl.GetError())
	}

	if sdl.NumJoysticks() > 0 {
		if controller := sdl.GameControllerOpen(0); controller != nil {
			defer controller.Close()
			fmt.Println("Controller detected.")
		}
	}

	currentSceneIndex = 0
	selectedButtonIndex = 0
	var inputText = ""

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.KeyboardEvent: // Use pointer receiver
				if e.Type == sdl.KEYDOWN {
					switch e.Keysym.Sym {
					case sdl.K_UP, sdl.K_w:
						moveSelection(config, -1)
					case sdl.K_DOWN, sdl.K_s:
						moveSelection(config, 1)
					case sdl.K_RETURN, sdl.K_KP_ENTER:
						triggerSelectedElement(config)
					case sdl.K_ESCAPE:
						running = false
					}
				}
			case *sdl.TextInputEvent: // Use pointer receiver
				inputText += string(e.Text[:])
			case *sdl.QuitEvent: // Use pointer receiver
				running = false
			case *sdl.MouseButtonEvent: // Use pointer receiver
				fmt.Printf("Mouse event")
				if e.Button == sdl.BUTTON_LEFT {
					mouseX := int(e.X)
					mouseY := int(e.Y)

					currentScene := config.Scenes[currentSceneIndex]
					for _, element := range currentScene.Elements {
						if element.Type == "button" &&
							mouseX > int(element.X)-50 && mouseX < int(element.X)+50 &&
							mouseY > int(element.Y)-25 && mouseY < int(element.Y)+25 {
							fmt.Println("Button clicked:", element.Text)
							handleTrigger(config, element)
						}
					}
				}
			case *sdl.ControllerButtonEvent: // Use pointer receiver
				if e.Type == sdl.CONTROLLERBUTTONDOWN {
					switch e.Button {
					case sdl.CONTROLLER_BUTTON_DPAD_UP:
						moveSelection(config, -1)
					case sdl.CONTROLLER_BUTTON_DPAD_DOWN:
						moveSelection(config, 1)
					case sdl.CONTROLLER_BUTTON_A:
						triggerSelectedElement(config)
					}
				}
			default:
				fmt.Printf("Unhandled event: %T\n", event)
			}
		}

		renderScene(renderer, config, config.Scenes[currentSceneIndex])
		renderer.Present()
		sdl.Delay(16)
	}
}

func moveSelection(config *Config, direction int) {
	elements := config.Scenes[currentSceneIndex].Elements
	numElements := len(elements)

	if numElements == 0 {
		return
	}

	selectedButtonIndex = (selectedButtonIndex + direction + numElements) % numElements
	for elements[selectedButtonIndex].Type != "button" {
		selectedButtonIndex = (selectedButtonIndex + direction + numElements) % numElements
	}
	fmt.Println("Selected button index:", selectedButtonIndex)
}

func triggerSelectedElement(config *Config) {
	selectedElement := config.Scenes[currentSceneIndex].Elements[selectedButtonIndex]
	if selectedElement.Type == "button" {
		fmt.Println("Triggering selected button:", selectedElement.Text)
		handleTrigger(config, selectedElement)
	}
}

func handleTrigger(config *Config, element Element) {
	if element.Trigger == "" {
		return
	}

	fmt.Println("Handling trigger for element:", element.Text)
	switch element.Trigger {
	case "exit":
		fmt.Println("Exiting the application.")
		os.Exit(0)
	case "change_scene":
		if element.TriggerTarget != "" {
			fmt.Println("Changing scene to:", element.TriggerTarget)
			for i, scene := range config.Scenes {
				if scene.Name == element.TriggerTarget {
					currentSceneIndex = i
					selectedButtonIndex = 0
					fmt.Println("Scene changed to:", element.TriggerTarget)
					break
				}
			}
		}
	case "set_variable":
		if element.TriggerTarget != "" && element.TriggerValue != "" {
			fmt.Println("Setting variable", element.TriggerTarget, "to", element.TriggerValue)
			setConfigVariable(config, element.TriggerTarget, element.TriggerValue)
		}
	}
}

func setConfigVariable(config *Config, variableName, value string) {
	fmt.Println("Setting configuration variable:", variableName, "to", value)
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
