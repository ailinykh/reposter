package xui

import (
	"fmt"
	"runtime"
	"slices"
)

var dict = map[string]string{
	"vpn_button_create_key": "🔑 Создать новый ключ",
	"vpn_button_manage_key": "🔐 Управление ключами",
	"vpn_button_remove_key": "❌ Удалить ключ",
	"vpn_button_back":       "⏪ Назад",
	"vpn_button_cancel":     "❌ Отмена",
	"vpn_enter_create_key_name": `Придумайте <b>любое имя</b> для ключа.
	
	Например:
	<i>- iPhone</i>
	<i>- Мой ключ</i>
	
	напишите имя в следующем сообщении`,
	"vpn_enter_create_key_name_too_long": "Давайте придумаем что-то более лаконичное",
	"vpn_enter_delete_key_name_top":      "Введите имя ключа, который хотите <b>удалить</b>\n",
	"vpn_enter_delete_key_name_item":     "<code>%s</code>",
	"vpn_key_created":                    "✅ Вы успешно создали новый ключ\n\n<code>%s</code>\n\nтеперь скопируйте ключ в буффер обмена (простым нажатием на него) и вставьте его в приложение",
	"vpn_key_deleted":                    "✅ Ключ \"<i>%s</i>\" удалён!\n\n",
	"vpn_key_not_found":                  "❌ Ключ не найден\n\n",
	"vpn_key_list_top":                   "🔑 Активные ключи:\n",
	"vpn_key_list_item":                  "<b>%d.</b> %s\n<code>%s</code>\n",
	"vpn_key_list_bottom":                "\nВсего ключей: <b>%d</b>",
	"vpn_mislead":                        "Неизвестная команда",
	"vpn_unexpected_state":               "Возникла непредвиденная ошибка, попробуйте начать сначала\n\n/vpnhelp",
	"vpn_welcome": `🌏 <b>VPN всего за 3 простых шага</b>
	
	1️⃣ Установите любой <b>vless-совместимый</b> клиент на ваше устройство, например:
	
	🍏 <a href='https://apps.apple.com/us/app/v2raytun/id6476628951'>v2RayTun</a> или <a href='https://apps.apple.com/ru/app/streisand/id6450534064?l=ru-RU'>Streisand</a> для iOS
	🤖 <a href='https://play.google.com/store/apps/details?id=com.v2raytun.android&hl=en'>v2RayTun</a> или <a href='https://play.google.com/store/apps/details?id=com.v2ray.vless&hl=en'>Vless VPN</a> для Android
	🖥️ <a href='https://apps.apple.com/ru/app/v2raytun/id6476628951?l=en-GB'>v2RayTun</a> для macOS
	 
	2️⃣ Нажмите на кнопку <i>"Создать новый ключ"</i> и следуйте инструкциям
	
	3️⃣ Скопируйте полученный ключ в клиент`,
}

func i18n(key string, args ...any) string {
	if val, ok := dict[key]; ok {
		return fmt.Sprintf(val, args...)
	}

	_, file, line, _ := runtime.Caller(0)
	return fmt.Sprintf("%s:%d KEY_MISSED:\"%s\"", file, line, key)
}

func allKeys() []string {
	keys := make([]string, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
