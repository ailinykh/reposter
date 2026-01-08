package telegram

func NewInlineKeyboard(keyboard [][]map[string]any) map[string]any {
	return map[string]any{
		"reply_markup": map[string]any{
			"inline_keyboard": keyboard,
		},
	}
}

func NewKeyboard(keyboard [][]map[string]any) map[string]any {
	if len(keyboard) == 0 {
		return map[string]any{
			"reply_markup": map[string]any{
				"remove_keyboard": true,
			},
		}
	}
	return map[string]any{
		"reply_markup": map[string]any{
			"keyboard": keyboard,
		},
	}
}
