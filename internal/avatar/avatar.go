package avatar

import (
	"fmt"
	"hash/fnv"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/gofiber/fiber/v2"
)

// Handler returns an SVG avatar with a gradient background keyed off the email.
func Handler(c *fiber.Ctx) error {
	email := c.Params("email")
	if strings.TrimSpace(email) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email parameter is required",
		})
	}

	size := parseAvatarSize(c.Query("size"))
	primary := avatarColor(email)
	secondary := lightenHexColor(primary, 0.25)
	letter := avatarInitial(email)

	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 128 128">`+
		`<defs><linearGradient id="avatar-grad" x1="0%%" y1="0%%" x2="100%%" y2="100%%">`+
		`<stop offset="0%%" stop-color="%s"/><stop offset="100%%" stop-color="%s"/>`+
		`</linearGradient></defs><rect width="128" height="128" rx="32" fill="url(#avatar-grad)"/>`+
		`<text x="50%%" y="58%%" text-anchor="middle" dominant-baseline="middle" font-size="64" fill="#ffffff" font-weight="600" font-family="Inter, system-ui, sans-serif">%s</text>`+
		`</svg>`, size, size, primary, secondary, letter)

	c.Set("Content-Type", "image/svg+xml; charset=utf-8")
	c.Set("Cache-Control", "public, max-age=86400, immutable")
	return c.SendString(svg)
}

func parseAvatarSize(query string) int {
	if query == "" {
		return 128
	}
	if size, err := strconv.Atoi(query); err == nil {
		if size < 48 {
			return 48
		}
		if size > 256 {
			return 256
		}
		return size
	}
	return 128
}

func avatarColor(seed string) string {
	colors := []string{"#ff6b35", "#f97316", "#0ea5e9", "#8b5cf6", "#ec4899", "#14b8a6", "#22c55e", "#facc15"}
	h := fnv.New32a()
	h.Write([]byte(seed))
	return colors[int(h.Sum32())%len(colors)]
}

func avatarInitial(value string) string {
	for _, r := range value {
		if unicode.IsLetter(r) {
			return strings.ToUpper(string(r))
		}
		if unicode.IsDigit(r) {
			return string(r)
		}
	}
	return "?"
}

func lightenHexColor(hex string, amount float64) string {
	if len(hex) != 7 || hex[0] != '#' {
		return hex
	}
	r, err := strconv.ParseInt(hex[1:3], 16, 64)
	if err != nil {
		return hex
	}
	g, err := strconv.ParseInt(hex[3:5], 16, 64)
	if err != nil {
		return hex
	}
	b, err := strconv.ParseInt(hex[5:7], 16, 64)
	if err != nil {
		return hex
	}
	return fmt.Sprintf("#%02x%02x%02x", clampComponent(r, amount), clampComponent(g, amount), clampComponent(b, amount))
}

func clampComponent(value int64, amount float64) int64 {
	max := 255.0
	return int64(math.Round(float64(value) + (max-float64(value))*amount))
}
