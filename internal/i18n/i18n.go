// Package i18n provides the minimal language support for the overlay.
//
// Only two display languages exist: English ("en") and Simplified Chinese
// ("zh-CN"). The user setting has a third value, "auto", which follows the
// host OS locale. Anything unrecognised - an unknown locale, an empty
// string, a missing dictionary key - falls back to English, never to a
// blank label.
package i18n

import (
	"strings"
	"syscall"
	"unsafe"
)

// Lang is a display language the overlay can render.
type Lang string

const (
	LangEn   Lang = "en"
	LangZhCN Lang = "zh-CN"
)

// Setting is the persisted language preference: follow the system or pin
// one of the two display languages. The zero value is meaningless on
// purpose - callers must go through NormalizeSetting.
type Setting string

const (
	SettingAuto Setting = "auto"
	SettingEn   Setting = "en"
	SettingZhCN Setting = "zh-CN"
)

// NormalizeSetting maps any stored value onto a known setting. Old or
// hand-edited config files may contain anything; they get "auto".
func NormalizeSetting(value string) Setting {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "en":
		return SettingEn
	case "zh-cn":
		return SettingZhCN
	default:
		return SettingAuto
	}
}

// NormalizeLang maps any value onto a display language, defaulting to
// English. It is the last line of defence before a label reaches the
// screen.
func NormalizeLang(value string) Lang {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(value)), "zh") {
		return LangZhCN
	}
	return LangEn
}

// Resolve picks the display language for a setting + system locale pair.
// "auto" follows the OS; explicit settings win; everything unknown is
// English.
func Resolve(setting Setting, systemLocale string) Lang {
	switch NormalizeSetting(string(setting)) {
	case SettingEn:
		return LangEn
	case SettingZhCN:
		return LangZhCN
	default:
		return NormalizeLang(systemLocale)
	}
}

// SystemLocale returns the host OS locale name ("zh-CN", "en-US", ...)
// or "" when it cannot be determined. Callers treat "" as English via
// NormalizeLang.
func SystemLocale() string {
	name, err := systemLocaleWindows()
	if err != nil || name == "" {
		return ""
	}
	return name
}

var (
	modKernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procGetUserDefaultLocaleName = modKernel32.NewProc("GetUserDefaultLocaleName")
)

// systemLocaleWindows reads the user default locale via the Win32 API.
// LOCALE_NAME_MAX_LENGTH is 85; anything longer is truncated, which is
// still enough to detect the language prefix we care about.
func systemLocaleWindows() (string, error) {
	const maxLen = 85
	buf := make([]uint16, maxLen)
	ret, _, err := procGetUserDefaultLocaleName.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(maxLen),
	)
	if ret == 0 {
		if err != nil && err != syscall.Errno(0) {
			return "", err
		}
		return "", syscall.EINVAL
	}
	return syscall.UTF16ToString(buf), nil
}

// Keys for T. New UI copy must reuse these instead of adding literals.
const (
	KeyStateWaiting  = "state.waiting"
	KeyStateError    = "state.error"
	KeyStateWorking  = "state.working"
	KeyStateTool     = "state.tool"
	KeyStateThinking = "state.thinking"
	KeyStateIdle     = "state.idle"

	KeyProjectUnknown = "project.unknown"

	KeyPanelSessions = "panel.sessions"
	KeyPanelTitle    = "panel.title"
	KeyPanelWaiting  = "panel.waiting"
	KeyPanelMore     = "panel.more"
	KeyPanelHint     = "panel.hint"
	KeyPanelEmpty    = "panel.empty"
	KeyPanelJump     = "panel.jump"

	KeySummaryEmpty    = "summary.empty"
	KeySummarySessions = "summary.sessions"
	KeySummaryRunning  = "summary.running"
	KeySummaryWaiting  = "summary.waiting"

	KeyTimeNow   = "time.now"
	KeyTimeMin1  = "time.min1"
	KeyTimeMins  = "time.mins"
	KeyTimeHour1 = "time.hour1"
	KeyTimeHours = "time.hours"

	// The short forms exist for the compact strip only: it is 56 logical units
	// wide, where "12 minutes ago" cannot be drawn. The expanded rows keep the
	// long wording, because there the sentence is the point.
	KeyTimeShortMin1  = "time.shortMin1"
	KeyTimeShortMins  = "time.shortMins"
	KeyTimeShortHour1 = "time.shortHour1"
	KeyTimeShortHours = "time.shortHours"

	KeyTrayShow      = "tray.show"
	KeyTrayHide      = "tray.hide"
	KeyTrayReset     = "tray.reset"
	KeyTrayNotify    = "tray.notify"
	KeyTraySound     = "tray.sound"
	KeyTrayTopmost   = "tray.topmost"
	KeyTrayRetract   = "tray.retract"
	KeyTrayAutostart = "tray.autostart"
	KeyTrayQuit      = "tray.quit"
	KeyTrayLangTitle = "tray.lang.title"
	KeyTrayLangAuto  = "tray.lang.auto"

	KeyTrayAutoApprove  = "tray.autoApprove"
	KeyTraySoundVolume  = "tray.soundVolume"
	KeyTrayNotifyFinish = "tray.notifyFinish"

	KeyBalloonWaitTitle = "balloon.waitTitle"
	KeyBalloonErrTitle  = "balloon.errTitle"
	KeyBalloonDoneTitle = "balloon.doneTitle"
	KeyBalloonDoneBody  = "balloon.doneBody"
	KeyBalloonJumpTitle = "balloon.jumpTitle"
	KeyBalloonJumpBody  = "balloon.jumpBody"

	KeyApproveCountdown = "approve.countdown"
	KeyApproveDenied    = "approve.denied"

	// Settings window. The window is a second top-level window, so its labels
	// live here rather than being assembled from tray strings: the tray needs
	// terse one-liners, the window has room for a hint under each row.
	KeyTraySettings = "tray.settings"

	KeySettingsTitle         = "settings.title"
	KeySettingsSubtitle      = "settings.subtitle"
	KeySettingsTabGeneral    = "settings.tab.general"
	KeySettingsTabSound      = "settings.tab.sound"
	KeySettingsTabSessions   = "settings.tab.sessions"
	KeySettingsTabAppearance = "settings.tab.appearance"
	KeySettingsClose         = "settings.close"
	KeySettingsOn            = "settings.on"
	KeySettingsOff           = "settings.off"

	KeySettingsNotify        = "settings.notify"
	KeySettingsNotifyHint    = "settings.notify.hint"
	KeySettingsTopmost       = "settings.topmost"
	KeySettingsTopmostHint   = "settings.topmost.hint"
	KeySettingsRetract       = "settings.retract"
	KeySettingsRetractHint   = "settings.retract.hint"
	KeySettingsOpacity       = "settings.opacity"
	KeySettingsOpacityHint   = "settings.opacity.hint"
	KeySettingsLanguage      = "settings.language"
	KeySettingsLanguageHint  = "settings.language.hint"
	KeySettingsAutostart     = "settings.autostart"
	KeySettingsAutostartHint = "settings.autostart.hint"

	KeySettingsSound            = "settings.sound"
	KeySettingsSoundHint        = "settings.sound.hint"
	KeySettingsScheme           = "settings.scheme"
	KeySettingsSchemeHint       = "settings.scheme.hint"
	KeySettingsVolume           = "settings.volume"
	KeySettingsVolumeHint       = "settings.volume.hint"
	KeySettingsSoundWaiting     = "settings.sound.waiting"
	KeySettingsSoundWaitingHint = "settings.sound.waiting.hint"
	KeySettingsSoundError       = "settings.sound.error"
	KeySettingsSoundErrorHint   = "settings.sound.error.hint"
	KeySettingsSoundFinish      = "settings.sound.finish"
	KeySettingsSoundFinishHint  = "settings.sound.finish.hint"
	KeySettingsPreview          = "settings.preview"
	KeySettingsPreviewHint      = "settings.preview.hint"
	KeySettingsPreviewPlay      = "settings.preview.play"

	KeySettingsTheme              = "settings.theme"
	KeySettingsThemeHint          = "settings.theme.hint"
	KeySettingsPerSource          = "settings.perSource"
	KeySettingsPerSourceHint      = "settings.perSource.hint"
	KeySettingsHiddenAtLaunch     = "settings.hiddenAtLaunch"
	KeySettingsHiddenAtLaunchHint = "settings.hiddenAtLaunch.hint"
	KeySettingsAppearanceNote     = "settings.appearance.note"

	KeyThemeSlate    = "theme.slate"
	KeyThemePaper    = "theme.paper"
	KeyThemeContrast = "theme.contrast"

	KeySettingsAutoApprove        = "settings.autoApprove"
	KeySettingsAutoApproveHint    = "settings.autoApprove.hint"
	KeySettingsApproveTimeout     = "settings.approveTimeout"
	KeySettingsApproveTimeoutHint = "settings.approveTimeout.hint"
	KeySettingsReset              = "settings.reset"
	KeySettingsResetHint          = "settings.reset.hint"
	KeySettingsResetAction        = "settings.reset.action"
	KeySettingsSessionsNote       = "settings.sessions.note"

	KeySchemeChime = "scheme.chime"
	KeySchemeBell  = "scheme.bell"
	KeySchemeSoft  = "scheme.soft"
	KeySchemePulse = "scheme.pulse"

	// Attentiveness: the settings that decide whether an alert is useful. They
	// sit on the Sessions tab next to auto-approve, because both answer "what
	// does this thing do behind my back".
	KeySettingsForegroundAware     = "settings.foregroundAware"
	KeySettingsForegroundAwareHint = "settings.foregroundAware.hint"
	KeySettingsIdleMinutes         = "settings.idleMinutes"
	KeySettingsIdleMinutesHint     = "settings.idleMinutes.hint"
	KeySettingsDonePolicy          = "settings.donePolicy"
	KeySettingsDonePolicyHint      = "settings.donePolicy.hint"
	KeySettingsDoneBackground      = "settings.donePolicy.background"
	KeySettingsDoneOff             = "settings.donePolicy.off"
	KeySettingsDoneAlways          = "settings.donePolicy.always"
	KeySettingsQuietCurrent        = "settings.quietCurrent"
	KeySettingsQuietCurrentHint    = "settings.quietCurrent.hint"
	KeySettingsQuietSubagents      = "settings.quietSubagents"
	KeySettingsQuietSubagentsHint  = "settings.quietSubagents.hint"
	KeySettingsMuteCount           = "settings.muteCount"
	KeySettingsMuteCountHint       = "settings.muteCount.hint"
	KeySettingsMuteClear           = "settings.muteClear"
	KeySettingsMuteClearAction     = "settings.muteClear.action"
	KeySettingsAttendNote          = "settings.attend.note"

	KeyTrayMuteSession   = "tray.muteSession"
	KeyTrayUnmuteSession = "tray.unmuteSession"

	KeyPanelUnread = "panel.unread"
	KeyPanelMuted  = "panel.muted"

	// Alert test buttons: the three things a user can be asked about, each
	// firing a real balloon through the same path a real alert uses.
	KeySettingsTestWaiting     = "settings.test.waiting"
	KeySettingsTestWaitingHint = "settings.test.waiting.hint"
	KeySettingsTestDone        = "settings.test.done"
	KeySettingsTestDoneHint    = "settings.test.done.hint"
	KeySettingsTestPlay        = "settings.test.play"

	KeyBalloonDoneNamed  = "balloon.doneNamed"
	KeyBalloonMutedTitle = "balloon.mutedTitle"
	KeyBalloonMutedBody  = "balloon.mutedBody"
	KeyAlertReceiptNote  = "alerts.receiptNote"
)

var dictEn = map[string]string{
	KeyStateWaiting:  "Waiting",
	KeyStateError:    "Error",
	KeyStateWorking:  "Working",
	KeyStateTool:     "Tool call",
	KeyStateThinking: "Thinking",
	KeyStateIdle:     "Idle",

	KeyProjectUnknown: "Untitled project",

	KeyPanelSessions: "sessions",
	KeyPanelTitle:    "Terminal sessions",
	KeyPanelWaiting:  "Waiting for AI CLI",
	KeyPanelMore:     "and %d more sessions",
	KeyPanelHint:     "Drag · Scroll to zoom · ↗ Jump · Click to collapse",
	KeyPanelEmpty:    "Launch an AI CLI in a terminal; the panel appears on its own",
	KeyPanelJump:     "Jump",

	KeySummaryEmpty:    "No sessions",
	KeySummarySessions: "%d sessions",
	KeySummaryRunning:  "%d running",
	KeySummaryWaiting:  "%d waiting",

	KeyTimeNow:   "just now",
	KeyTimeMin1:  "1 minute ago",
	KeyTimeMins:  "%d minutes ago",
	KeyTimeHour1: "1 hour ago",
	KeyTimeHours: "%d hours ago",

	KeyTimeShortMin1:  "1m",
	KeyTimeShortMins:  "%dm",
	KeyTimeShortHour1: "1h",
	KeyTimeShortHours: "%dh",

	KeyTrayShow:      "Show panel",
	KeyTrayHide:      "Hide to tray",
	KeyTrayReset:     "Reset position and size",
	KeyTrayNotify:    "Status alerts",
	KeyTraySound:     "Alert sound",
	KeyTrayTopmost:   "Always on top",
	KeyTrayRetract:   "Dock and retract",
	KeyTrayAutostart: "Run at logon",
	KeyTrayQuit:      "Quit",
	KeyTrayLangTitle: "Language",
	KeyTrayLangAuto:  "Auto (System)",

	KeyTrayAutoApprove:  "Auto-approve tool permissions after 5s",
	KeyTraySoundVolume:  "Alert volume",
	KeyTrayNotifyFinish: "Alert when a session finishes",

	KeyBalloonWaitTitle: "Waiting for your approval",
	KeyBalloonErrTitle:  "Terminal session failed",
	KeyBalloonDoneTitle: "Session finished",
	KeyBalloonDoneBody:  "The task is done - switch back when you are ready",

	KeyBalloonJumpTitle: "Could not switch terminal",
	KeyBalloonJumpBody:  "No matching terminal window is open for this session",

	KeyApproveCountdown: "Allow (%ds)",
	KeyApproveDenied:    "Denied",

	KeyTraySettings: "Settings…",

	KeySettingsTitle:         "Settings",
	KeySettingsSubtitle:      "AI Terminal Pet",
	KeySettingsTabGeneral:    "General",
	KeySettingsTabSound:      "Sound",
	KeySettingsTabSessions:   "Sessions",
	KeySettingsTabAppearance: "Look",
	KeySettingsClose:         "Close",
	KeySettingsOn:            "On",
	KeySettingsOff:           "Off",

	KeySettingsNotify:        "Status alerts",
	KeySettingsNotifyHint:    "Balloon when a session waits or fails",
	KeySettingsTopmost:       "Always on top",
	KeySettingsTopmostHint:   "Keep the panel above other windows",
	KeySettingsRetract:       "Dock and retract",
	KeySettingsRetractHint:   "Slide into the screen edge when idle",
	KeySettingsOpacity:       "Panel opacity",
	KeySettingsOpacityHint:   "Lower it to make the panel recede",
	KeySettingsLanguage:      "Language",
	KeySettingsLanguageHint:  "Auto follows the system locale",
	KeySettingsAutostart:     "Run at logon",
	KeySettingsAutostartHint: "Start the panel with Windows",

	KeySettingsSound:            "Alert sound",
	KeySettingsSoundHint:        "Synthesised tones, no external files",
	KeySettingsScheme:           "Tone",
	KeySettingsSchemeHint:       "Click a tone to hear it",
	KeySettingsVolume:           "Volume",
	KeySettingsVolumeHint:       "Applied to every alert tone",
	KeySettingsSoundWaiting:     "Waiting for approval",
	KeySettingsSoundWaitingHint: "A session is blocked on you",
	KeySettingsSoundError:       "Session failed",
	KeySettingsSoundErrorHint:   "A session stopped with an error",
	KeySettingsSoundFinish:      "Session finished",
	KeySettingsSoundFinishHint:  "A busy session settled down",
	KeySettingsPreview:          "Audition",
	KeySettingsPreviewHint:      "Play the tone selected above",
	KeySettingsPreviewPlay:      "Play",

	KeySettingsTheme:              "Theme",
	KeySettingsThemeHint:          "Palette for the panel and this window",
	KeySettingsPerSource:          "Colour by terminal",
	KeySettingsPerSourceHint:      "Tint each row's source name",
	KeySettingsHiddenAtLaunch:     "Start hidden",
	KeySettingsHiddenAtLaunchHint: "Begin in the tray; a reporter brings it back",
	KeySettingsAppearanceNote:     "The palette applies to the panel and to this window. Every screenshot in the repository uses Slate.",

	KeyThemeSlate:    "Slate",
	KeyThemePaper:    "Paper",
	KeyThemeContrast: "Contrast",

	KeySettingsAutoApprove:        "Auto-approve permissions",
	KeySettingsAutoApproveHint:    "Approve when the countdown runs out",
	KeySettingsApproveTimeout:     "Countdown",
	KeySettingsApproveTimeoutHint: "Seconds before a prompt auto-approves",
	KeySettingsReset:              "Restore defaults",
	KeySettingsResetHint:          "Does not touch window position",
	KeySettingsResetAction:        "Restore",
	KeySettingsSessionsNote:       "Auto-approve only works for terminals that report permission prompts. Leave it off unless you are watching.",

	KeySchemeChime: "Chime",
	KeySchemeBell:  "Bell",
	KeySchemeSoft:  "Soft",
	KeySchemePulse: "Pulse",

	KeySettingsForegroundAware:     "Only alert when I am not looking",
	KeySettingsForegroundAwareHint: "A prompt you are reading does not also pop a balloon; its row is still marked unread",
	KeySettingsIdleMinutes:         "Count me away after",
	KeySettingsIdleMinutesHint:     "Longest quiet stretch before a window in front stops counting as attention; 0 means window focus only",
	KeySettingsDonePolicy:          "When a reply finishes",
	KeySettingsDonePolicyHint:      "Background: only sessions you were not watching; off: never; always: every time",
	KeySettingsDoneBackground:      "Background",
	KeySettingsDoneOff:             "Off",
	KeySettingsDoneAlways:          "Always",
	KeySettingsQuietCurrent:        "Stay quiet for the session I am reading",
	KeySettingsQuietCurrentHint:    "Same session, no balloon, unread mark instead; errors and tones are unaffected",
	KeySettingsQuietSubagents:      "Silence child sessions",
	KeySettingsQuietSubagentsHint:  "Sessions that declare a parent are steps of one turn, not separate work",
	KeySettingsMuteCount:           "Muted sessions",
	KeySettingsMuteCountHint:       "Right-click a session row to mute or unmute it; this clears the whole list",
	KeySettingsMuteClear:           "Mute list",
	KeySettingsMuteClearAction:     "Clear all",
	KeySettingsAttendNote:          "These decide whether an alert is useful, never whether a session is safe: auto-approve still counts down while muted.",

	KeyTrayMuteSession:   "Mute this session's alerts",
	KeyTrayUnmuteSession: "Unmute this session's alerts",

	KeyPanelUnread: "unread",
	KeyPanelMuted:  "muted",

	KeySettingsTestWaiting:     "Send a test: needs you",
	KeySettingsTestWaitingHint: "A real balloon and tone through the same path an approval uses; counts against no session",
	KeySettingsTestDone:        "Send a test: reply finished",
	KeySettingsTestDoneHint:    "The completion balloon with the timbre of the current Done setting",
	KeySettingsTestPlay:        "Send",

	KeyBalloonDoneNamed:  "Finished",
	KeyBalloonMutedTitle: "Session muted",
	KeyBalloonMutedBody:  "Its alerts are off until you unmute the row",
	KeyAlertReceiptNote:  "The last alert decision is written to runtime/alerts.json and printed by --json.",
}

var dictZhCN = map[string]string{
	KeyStateWaiting:  "等待确认",
	KeyStateError:    "错误",
	KeyStateWorking:  "执行中",
	KeyStateTool:     "工具调用",
	KeyStateThinking: "思考中",
	KeyStateIdle:     "空闲",

	KeyProjectUnknown: "未标注项目",

	KeyPanelSessions: "会话",
	KeyPanelTitle:    "终端会话",
	KeyPanelWaiting:  "等待 AI CLI",
	KeyPanelMore:     "还有 %d 个会话",
	KeyPanelHint:     "拖拽 · 滚轮缩放 · ↗ 跳转 · 点击收起",
	KeyPanelEmpty:    "在终端里启动 AI CLI，浮窗会自动出现",
	KeyPanelJump:     "跳转",

	KeySummaryEmpty:    "暂无会话",
	KeySummarySessions: "%d 个会话",
	KeySummaryRunning:  "%d 运行",
	KeySummaryWaiting:  "%d 等待",

	KeyTimeNow:   "刚刚",
	KeyTimeMin1:  "1 分钟前",
	KeyTimeMins:  "%d 分钟前",
	KeyTimeHour1: "1 小时前",
	KeyTimeHours: "%d 小时前",

	KeyTimeShortMin1:  "1分",
	KeyTimeShortMins:  "%d分",
	KeyTimeShortHour1: "1小时",
	KeyTimeShortHours: "%d小时",

	KeyTrayShow:      "显示浮窗",
	KeyTrayHide:      "隐藏到托盘",
	KeyTrayReset:     "重置位置与大小",
	KeyTrayNotify:    "状态提醒",
	KeyTraySound:     "提示音",
	KeyTrayTopmost:   "窗口置顶",
	KeyTrayRetract:   "贴边收起",
	KeyTrayAutostart: "开机自启",
	KeyTrayQuit:      "退出",
	KeyTrayLangTitle: "语言 / Language",
	KeyTrayLangAuto:  "自动（跟随系统）",

	KeyTrayAutoApprove:  "5秒未响应时自动允许权限请求",
	KeyTraySoundVolume:  "提示音音量",
	KeyTrayNotifyFinish: "会话结束时提醒",

	KeyBalloonWaitTitle: "等待你的确认",
	KeyBalloonErrTitle:  "终端会话出错",
	KeyBalloonDoneTitle: "会话已结束",
	KeyBalloonDoneBody:  "任务已完成，方便时切回终端查看结果",

	KeyBalloonJumpTitle: "无法跳转到终端",
	KeyBalloonJumpBody:  "没有找到这个会话对应的终端窗口",

	KeyApproveCountdown: "允许（%d 秒）",
	KeyApproveDenied:    "已拒绝",

	KeyTraySettings: "设置…",

	KeySettingsTitle:         "设置",
	KeySettingsSubtitle:      "AI Terminal Pet",
	KeySettingsTabGeneral:    "通用",
	KeySettingsTabSound:      "声音",
	KeySettingsTabSessions:   "会话",
	KeySettingsTabAppearance: "外观",
	KeySettingsClose:         "关闭",
	KeySettingsOn:            "开",
	KeySettingsOff:           "关",

	KeySettingsNotify:        "状态提醒",
	KeySettingsNotifyHint:    "会话等待确认或出错时弹出气泡",
	KeySettingsTopmost:       "窗口置顶",
	KeySettingsTopmostHint:   "让面板始终浮在其他窗口之上",
	KeySettingsRetract:       "贴边收起",
	KeySettingsRetractHint:   "空闲时滑入屏幕边缘，只留一条状态脊",
	KeySettingsOpacity:       "面板不透明度",
	KeySettingsOpacityHint:   "调低可以让面板更不打扰",
	KeySettingsLanguage:      "语言",
	KeySettingsLanguageHint:  "自动跟随系统区域设置",
	KeySettingsAutostart:     "开机自启",
	KeySettingsAutostartHint: "随 Windows 启动面板",

	KeySettingsSound:            "提示音",
	KeySettingsSoundHint:        "本地合成音，不依赖音频文件",
	KeySettingsScheme:           "音效方案",
	KeySettingsSchemeHint:       "点击任意方案即可试听",
	KeySettingsVolume:           "音量",
	KeySettingsVolumeHint:       "作用于全部提示音",
	KeySettingsSoundWaiting:     "等待确认时",
	KeySettingsSoundWaitingHint: "有会话正卡在等你处理",
	KeySettingsSoundError:       "会话出错时",
	KeySettingsSoundErrorHint:   "会话异常中止",
	KeySettingsSoundFinish:      "会话完成时",
	KeySettingsSoundFinishHint:  "忙碌的会话转入空闲",
	KeySettingsPreview:          "试听音效",
	KeySettingsPreviewHint:      "播放当前选中的音效",
	KeySettingsPreviewPlay:      "播放",

	KeySettingsTheme:              "配色主题",
	KeySettingsThemeHint:          "面板与本设置窗口共用的调色板",
	KeySettingsPerSource:          "按终端着色",
	KeySettingsPerSourceHint:      "给每行的来源名加上该终端的颜色",
	KeySettingsHiddenAtLaunch:     "启动时隐藏",
	KeySettingsHiddenAtLaunchHint: "启动后先进托盘，有会话上报再出现",
	KeySettingsAppearanceNote:     "调色板同时作用于面板与这个设置窗口。仓库里所有截图用的都是「石板」。",

	KeyThemeSlate:    "石板",
	KeyThemePaper:    "纸白",
	KeyThemeContrast: "高对比",

	KeySettingsAutoApprove:        "自动同意权限",
	KeySettingsAutoApproveHint:    "倒计时结束仍未响应时自动允许",
	KeySettingsApproveTimeout:     "倒计时时长",
	KeySettingsApproveTimeoutHint: "等待多少秒后自动同意",
	KeySettingsReset:              "恢复默认设置",
	KeySettingsResetHint:          "不会改动窗口位置",
	KeySettingsResetAction:        "恢复",
	KeySettingsSessionsNote:       "自动同意只对会上报权限请求的终端有效。除非你正在旁边看着，否则建议保持关闭。",

	KeySchemeChime: "清脆",
	KeySchemeBell:  "钟声",
	KeySchemeSoft:  "柔和",
	KeySchemePulse: "脉冲",

	KeySettingsForegroundAware:     "只在我不看着时提醒",
	KeySettingsForegroundAwareHint: "正在阅读的请求不再额外弹气泡，但是那一行仍会标为未读",
	KeySettingsIdleMinutes:         "静止多久算离开",
	KeySettingsIdleMinutesHint:     "窗口在前台但多久没有键鼠输入就不再算你在看；0 表示只看窗口焦点",
	KeySettingsDonePolicy:          "回复完成时",
	KeySettingsDonePolicyHint:      "仅后台：只提醒你没在看的会话；关闭：从不；总是：每次",
	KeySettingsDoneBackground:      "仅后台",
	KeySettingsDoneOff:             "关闭",
	KeySettingsDoneAlways:          "总是",
	KeySettingsQuietCurrent:        "正在看的会话保持安静",
	KeySettingsQuietCurrentHint:    "同一个会话不弹气泡、只留未读标记；错误与提示音不受影响",
	KeySettingsQuietSubagents:      "子会话不提醒",
	KeySettingsQuietSubagentsHint:  "声明了上级会话的那些是同一轮的步骤，不是独立工作",
	KeySettingsMuteCount:           "已静音的会话",
	KeySettingsMuteCountHint:       "右键点会话行可以静音或恢复；这里清空整个列表",
	KeySettingsMuteClear:           "静音列表",
	KeySettingsMuteClearAction:     "全部恢复",
	KeySettingsAttendNote:          "这些只影响提醒是否有用，不影响会话是否安全：静音期间自动同意倒计时照常进行。",

	KeyTrayMuteSession:   "静音此会话的提醒",
	KeyTrayUnmuteSession: "恢复此会话的提醒",

	KeyPanelUnread: "未读",
	KeyPanelMuted:  "已静音",

	KeySettingsTestWaiting:     "试发：需要你处理",
	KeySettingsTestWaitingHint: "走与真实审批相同的路径弹一条真气泡并响提示音；不挂到任何会话上",
	KeySettingsTestDone:        "试发：回复完成",
	KeySettingsTestDoneHint:    "用当前「回复完成」设置的音色发一条完成气泡",
	KeySettingsTestPlay:        "发送",

	KeyBalloonDoneNamed:  "已完成",
	KeyBalloonMutedTitle: "该会话已静音",
	KeyBalloonMutedBody:  "在行上恢复之前，它的提醒保持关闭",
	KeyAlertReceiptNote:  "最近的提醒判定写入 runtime/alerts.json，并由 --json 打印。",
}

// T returns the label for key in lang. A missing key falls back to
// English; a key missing there too returns the key itself so a gap is
// always visible instead of blank.
func T(lang Lang, key string) string {
	if lang == LangZhCN {
		if value, ok := dictZhCN[key]; ok {
			return value
		}
	}
	if value, ok := dictEn[key]; ok {
		return value
	}
	return key
}
