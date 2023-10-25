package useragent

import (
	"bytes"
	"regexp"
	"strings"
)

// UserAgent struct containing all data extracted from parsed user-agent string
type UserAgent struct {
	VersionNo   VersionNo
	OSVersionNo VersionNo
	URL         string
	String      string
	Name        string
	Version     string
	OS          string
	OSVersion   string
	Device      string
	Mobile      bool
	Tablet      bool
	Desktop     bool
	Bot         bool
}

// Constants for browsers and operating systems for easier comparison
const (
	Windows      = "Windows"
	WindowsPhone = "Windows Phone"
	Android      = "Android"
	MacOS        = "macOS"
	IOS          = "iOS"
	Linux        = "Linux"
	FreeBSD      = "FreeBSD"
	ChromeOS     = "ChromeOS"
	BlackBerry   = "BlackBerry"

	Opera            = "Opera"
	OperaMini        = "Opera Mini"
	OperaTouch       = "Opera Touch"
	Chrome           = "Chrome"
	HeadlessChrome   = "Headless Chrome"
	Firefox          = "Firefox"
	InternetExplorer = "Internet Explorer"
	Safari           = "Safari"
	Edge             = "Edge"
	Vivaldi          = "Vivaldi"

	GoogleAdsBot        = "Google Ads Bot"
	Googlebot           = "Googlebot"
	Twitterbot          = "Twitterbot"
	FacebookExternalHit = "facebookexternalhit"
	Applebot            = "Applebot"
	Bingbot             = "Bingbot"

	FacebookApp  = "Facebook App"
	InstagramApp = "Instagram App"
	TiktokApp    = "TikTok App"
)

// Parses parses user agents.
// It is not safe to use concurrently.
type Parser struct {
	tokens properties
	buff   bytes.Buffer
	val    bytes.Buffer
}

// New creates a user agent parser.
func New() *Parser {
	return &Parser{
		tokens: properties{
			list: make([]property, 0, 8),
		},
	}
}

// defaultParser is the default Parser used by Parse.
var defaultParser = New()

// Parse parses a user agent using the default parser.
// It is not safe to use concurrently.
func Parse(userAgent string) UserAgent {
	return defaultParser.Parse(userAgent)
}

// Parse parses a user agent.
// It is not safe to use concurrently.
func (p *Parser) Parse(userAgent string) UserAgent {
	ua := UserAgent{
		String: userAgent,
	}

	p.parse(userAgent)

	// check is there URL
	for i, token := range p.tokens.list {
		if strings.HasPrefix(token.Key, "http://") || strings.HasPrefix(token.Key, "https://") {
			ua.URL = token.Key
			p.tokens.list = append(p.tokens.list[:i], p.tokens.list[i+1:]...)
			break
		}
	}

	//fmt.Printf("%+v\n", tokens)

	// OS lookup
	switch {
	case p.tokens.exists("Android"):
		ua.OS = Android
		var osIndex int
		osIndex, ua.OSVersion = p.tokens.getIndexValue(Android)
		ua.Tablet = strings.Contains(strings.ToLower(ua.String), "tablet")
		ua.Device = p.tokens.findAndroidDevice(osIndex)

	case p.tokens.exists("iPhone"):
		ua.OS = IOS
		ua.OSVersion = p.tokens.findMacOSVersion()
		ua.Device = "iPhone"
		ua.Mobile = true

	case p.tokens.exists("iPad"):
		ua.OS = IOS
		ua.OSVersion = p.tokens.findMacOSVersion()
		ua.Device = "iPad"
		ua.Tablet = true

	case p.tokens.exists("Windows NT"):
		ua.OS = Windows
		ua.OSVersion = p.tokens.get("Windows NT")
		ua.Desktop = true

	case p.tokens.exists("Windows Phone OS"):
		ua.OS = WindowsPhone
		ua.OSVersion = p.tokens.get("Windows Phone OS")
		ua.Mobile = true

	case p.tokens.exists("Macintosh"):
		ua.OS = MacOS
		ua.OSVersion = p.tokens.findMacOSVersion()
		ua.Desktop = true

	case p.tokens.exists("Linux"):
		ua.OS = Linux
		ua.OSVersion = p.tokens.get(Linux)
		ua.Desktop = true

	case p.tokens.exists("FreeBSD"):
		ua.OS = FreeBSD
		ua.OSVersion = p.tokens.get(FreeBSD)
		ua.Desktop = true

	case p.tokens.exists("CrOS"):
		ua.OS = ChromeOS
		ua.OSVersion = p.tokens.get("CrOS")
		ua.Desktop = true

	case p.tokens.exists("BlackBerry"):
		ua.OS = BlackBerry
		ua.OSVersion = p.tokens.get("BlackBerry")
		ua.Mobile = true
	}

	switch {
	case p.tokens.exists("Googlebot"):
		ua.Name = Googlebot
		ua.Version = p.tokens.get(Googlebot)
		ua.Bot = true
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	case p.tokens.existsAny("GoogleProber", "GoogleProducer"):
		if name := p.tokens.findBestMatch(false); name != "" {
			ua.Name = name
		}
		ua.Bot = true

	case p.tokens.exists("Applebot"):
		ua.Name = Applebot
		ua.Version = p.tokens.get(Applebot)
		ua.Bot = true
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")
		ua.OS = ""

	case p.tokens.get("Opera Mini") != "":
		ua.Name = OperaMini
		ua.Version = p.tokens.get(OperaMini)
		ua.Mobile = true

	case p.tokens.get("OPR") != "":
		ua.Name = Opera
		ua.Version = p.tokens.get("OPR")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	case p.tokens.get("OPT") != "":
		ua.Name = OperaTouch
		ua.Version = p.tokens.get("OPT")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	// Opera on iOS
	case p.tokens.get("OPiOS") != "":
		ua.Name = Opera
		ua.Version = p.tokens.get("OPiOS")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	// Chrome on iOS
	case p.tokens.get("CriOS") != "":
		ua.Name = Chrome
		ua.Version = p.tokens.get("CriOS")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	// Firefox on iOS
	case p.tokens.get("FxiOS") != "":
		ua.Name = Firefox
		ua.Version = p.tokens.get("FxiOS")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	case p.tokens.get("Firefox") != "":
		ua.Name = Firefox
		ua.Version = p.tokens.get(Firefox)
		ua.Mobile = p.tokens.exists("Mobile")
		ua.Tablet = p.tokens.exists("Tablet")

	case p.tokens.get("Vivaldi") != "":
		ua.Name = Vivaldi
		ua.Version = p.tokens.get(Vivaldi)

	case p.tokens.exists("MSIE"):
		ua.Name = InternetExplorer
		ua.Version = p.tokens.get("MSIE")

	case p.tokens.get("EdgiOS") != "":
		ua.Name = Edge
		ua.Version = p.tokens.get("EdgiOS")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	case p.tokens.get("Edge") != "":
		ua.Name = Edge
		ua.Version = p.tokens.get("Edge")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	case p.tokens.get("Edg") != "":
		ua.Name = Edge
		ua.Version = p.tokens.get("Edg")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	case p.tokens.get("EdgA") != "":
		ua.Name = Edge
		ua.Version = p.tokens.get("EdgA")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	case p.tokens.get("bingbot") != "":
		ua.Name = Bingbot
		ua.Version = p.tokens.get("bingbot")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	case p.tokens.get("YandexBot") != "":
		ua.Name = "YandexBot"
		ua.Version = p.tokens.get("YandexBot")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	case p.tokens.get("SamsungBrowser") != "":
		ua.Name = "Samsung Browser"
		ua.Version = p.tokens.get("SamsungBrowser")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	case p.tokens.get("HeadlessChrome") != "":
		ua.Name = HeadlessChrome
		ua.Version = p.tokens.get("HeadlessChrome")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")
		ua.Bot = true

	case p.tokens.existsAny("AdsBot-Google-Mobile", "Mediapartners-Google", "AdsBot-Google"):
		ua.Name = GoogleAdsBot
		ua.Bot = true
		ua.Mobile = ua.IsAndroid() || ua.IsIOS()

	case p.tokens.exists("Yahoo Ad monitoring"):
		ua.Name = "Yahoo Ad monitoring"
		ua.Bot = true
		ua.Mobile = ua.IsAndroid() || ua.IsIOS()

	case p.tokens.exists("XiaoMi"):
		miui := p.tokens.get("XiaoMi")
		if strings.HasPrefix(miui, "MiuiBrowser") {
			ua.Name = "Miui Browser"
			ua.Version = strings.TrimPrefix(miui, "MiuiBrowser/")
			ua.Mobile = true
		}

	case p.tokens.exists("FBAN"):
		ua.Name = FacebookApp
		ua.Version = p.tokens.get("FBAN")

	case p.tokens.exists("FB_IAB"):
		ua.Name = FacebookApp
		ua.Version = p.tokens.get("FBAV")

	case p.tokens.startsWith("Instagram"):
		ua.Name = InstagramApp
		ua.Version = p.tokens.findInstagramVersion()

	case p.tokens.exists("BytedanceWebview"):
		ua.Name = TiktokApp
		ua.Version = p.tokens.get("app_version")

	case p.tokens.get("HuaweiBrowser") != "":
		ua.Name = "Huawei Browser"
		ua.Version = p.tokens.get("HuaweiBrowser")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	case p.tokens.exists("BlackBerry"):
		ua.Name = "BlackBerry"
		ua.Version = p.tokens.get("Version")

	case p.tokens.exists("NetFront"):
		ua.Name = "NetFront"
		ua.Version = p.tokens.get("NetFront")
		ua.Mobile = true

	// if chrome and Safari defined, find any other token sent descr
	case p.tokens.exists(Chrome) && p.tokens.exists(Safari):
		name := p.tokens.findBestMatch(true)
		if name != "" {
			ua.Name = name
			ua.Version = p.tokens.get(name)
			break
		}
		fallthrough

	case p.tokens.exists("Chrome"):
		ua.Name = Chrome
		ua.Version = p.tokens.get("Chrome")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	case p.tokens.exists("Brave Chrome"):
		ua.Name = Chrome
		ua.Version = p.tokens.get("Brave Chrome")
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	case p.tokens.exists("Safari"):
		ua.Name = Safari
		v := p.tokens.get("Version")
		if v != "" {
			ua.Version = v
		} else {
			ua.Version = p.tokens.get("Safari")
		}
		ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")

	default:
		if ua.OS == "Android" && p.tokens.get("Version") != "" {
			ua.Name = "Android browser"
			ua.Version = p.tokens.get("Version")
			ua.Mobile = true
		} else {
			if name := p.tokens.findBestMatch(false); name != "" {
				ua.Name = name
				ua.Version = p.tokens.get(name)
			} else {
				ua.Name = ua.String
			}
			ua.Bot = strings.Contains(strings.ToLower(ua.Name), "bot")
			// If mobile flag has already been set, don't override it.
			if !ua.Mobile {
				ua.Mobile = p.tokens.existsAny("Mobile", "Mobile Safari")
			}
		}
	}

	if ua.IsAndroid() {
		ua.Mobile = true
	}

	// if tablet, switch mobile to off
	if ua.Tablet {
		ua.Mobile = false
	}

	// if not already bot, check some popular bots and wether URL is set
	if !ua.Bot {
		ua.Bot = ua.URL != ""
	}

	if !ua.Bot {
		switch ua.Name {
		case Twitterbot, FacebookExternalHit:
			ua.Bot = true
		}
	}

	parseVersion(ua.Version, &ua.VersionNo)
	parseVersion(ua.OSVersion, &ua.OSVersionNo)

	return ua
}

func (p *Parser) parse(userAgent string) {
	p.tokens.list = p.tokens.list[:0]

	p.buff.Reset()
	p.val.Reset()
	slash := false
	isURL := false

	addToken := func() {
		if p.buff.Len() != 0 {
			s := strings.TrimSpace(p.buff.String())
			if !ignore(s) {
				if isURL {
					s = strings.TrimPrefix(s, "+")
				}

				if p.val.Len() == 0 { // only if value don't exists
					var ver string
					s, ver = checkVer(s) // determin version string and split
					p.tokens.add(s, ver)
				} else {
					p.tokens.add(s, strings.TrimSpace(p.val.String()))
				}
			}
		}
		p.buff.Reset()
		p.val.Reset()
		slash = false
		isURL = false
	}

	parOpen := false
	braOpen := false

	bua := []byte(userAgent)
	for i, c := range bua {

		//fmt.Println(string(c), c)
		switch {
		case c == 41: // )
			addToken()
			parOpen = false

		case (parOpen || braOpen) && c == 59: // ;
			addToken()

		case c == 59: // ;
			addToken()

		case c == 40: // (
			addToken()
			parOpen = true

		case c == 91: // [
			addToken()
			braOpen = true
		case c == 93: // ]
			addToken()
			braOpen = false

		case c == 58: // :
			if bytes.HasSuffix(p.buff.Bytes(), []byte("http")) || bytes.HasSuffix(p.buff.Bytes(), []byte("https")) {
				// If we are part of a URL just write the character.
				p.buff.WriteByte(c)
			} else if i != len(bua)-1 && bua[i+1] != ' ' {
				// If the following character is not a space, change to a space.
				p.buff.WriteByte(' ')
			}
			// Otherwise don't write as its probably a badly formatted key value separator.

		case slash && c == 32:
			addToken()

		case slash:
			p.val.WriteByte(c)

		case c == 47 && !isURL: //   /
			if i != len(bua)-1 && bua[i+1] == 47 && (bytes.HasSuffix(p.buff.Bytes(), []byte("http:")) || bytes.HasSuffix(p.buff.Bytes(), []byte("https:"))) {
				p.buff.WriteByte(c)
				isURL = true
			} else {
				if ignore(p.buff.String()) {
					p.buff.Reset()
				} else {
					slash = true
				}
			}

		default:
			p.buff.WriteByte(c)
		}
	}
	addToken()
}

func checkVer(s string) (name, v string) {
	i := strings.LastIndex(s, " ")
	if i == -1 {
		return s, ""
	}

	//v = s[i+1:]

	switch s[:i] {
	case "Linux", "Windows NT", "Windows Phone OS", "MSIE", "Android":
		return s[:i], s[i+1:]
	case "CrOS x86_64", "CrOS aarch64", "CrOS armv7l":
		j := strings.LastIndex(s[:i], " ")
		return s[:j], s[j+1 : i]
	default:
		return s, ""
	}

	// for _, c := range v {
	// 	if (c >= 48 && c <= 57) || c == 46 {
	// 	} else {
	// 		return s, ""
	// 	}
	// }
	// return s[:i], s[i+1:]
}

// ignore retursn true if token should be ignored
func ignore(s string) bool {
	switch s {
	case "KHTML, like Gecko", "U", "compatible", "Mozilla", "WOW64", "en", "en-us", "en-gb", "ru-ru", "Browser":
		return true
	default:
		return false
	}
}

type property struct {
	Key   string
	Value string
}
type properties struct {
	list []property
}

func (p *properties) add(key, value string) {
	p.list = append(p.list, property{Key: key, Value: value})
}

func (p *properties) get(key string) string {
	for _, prop := range p.list {
		if prop.Key == key {
			return prop.Value
		}
	}
	return ""
}

func (p *properties) getIndexValue(key string) (int, string) {
	for i, prop := range p.list {
		if prop.Key == key {
			return i, prop.Value
		}
	}
	return -1, ""
}

func (p *properties) exists(key string) bool {
	for _, prop := range p.list {
		if prop.Key == key {
			return true
		}
	}
	return false
}

// func (p *properties) existsIgnoreCase(key string) bool {
// 	for _, prop := range p.list {
// 		if strings.EqualFold(prop.Key, key) {
// 			return true
// 		}
// 	}
// 	return false
// }

func (p *properties) existsAny(keys ...string) bool {
	for _, k := range keys {
		for _, prop := range p.list {
			if prop.Key == k {
				return true
			}
		}
	}
	return false
}

func (p *properties) findMacOSVersion() string {
	for _, token := range p.list {
		if strings.Contains(token.Key, "OS") {
			if ver := findVersion(token.Value); ver != "" {
				return ver
			} else if ver = findVersion(token.Key); ver != "" {
				return ver
			}
		}

	}
	return ""
}

func (p *properties) startsWith(value string) bool {
	for _, prop := range p.list {
		if strings.HasPrefix(prop.Key, value) {
			return true
		}
	}
	return false
}

func (p *properties) findInstagramVersion() string {
	for _, token := range p.list {
		if strings.HasPrefix(token.Key, "Instagram") {
			if ver := findVersion(token.Value); ver != "" {
				return ver
			} else if ver = findVersion(token.Key); ver != "" {
				return ver
			}
		}

	}
	return ""
}

// findBestMatch from the rest of the bunch
// in first cycle only return key with version value
// if withVerValue is false, do another cycle and return any token
func (p *properties) findBestMatch(withVerOnly bool) string {
	n := 2
	if withVerOnly {
		n = 1
	}
	for i := 0; i < n; i++ {
		for _, prop := range p.list {
			switch prop.Key {
			case Chrome, Firefox, Safari, "Version", "Mobile", "Mobile Safari", "Mozilla", "AppleWebKit", "Windows NT", "Windows Phone OS", Android, "Macintosh", Linux, "GSA", "CrOS", "Tablet":
			default:
				// don' pick if starts with number
				if len(prop.Key) != 0 && prop.Key[0] >= 48 && prop.Key[0] <= 57 {
					break
				}
				if i == 0 {
					if prop.Value != "" { // in first check, only return keys with value
						return prop.Key
					}
				} else {
					return prop.Key
				}
			}
		}
	}
	return ""
}

var rxMacOSVer = regexp.MustCompile(`[_\d\.]+`)

func findVersion(s string) string {
	if ver := rxMacOSVer.FindString(s); ver != "" {
		return strings.Replace(ver, "_", ".", -1)
	}
	return ""
}

// findAndroidDevice in tokens
func (p *properties) findAndroidDevice(startIndex int) string {
	for i := startIndex; i < startIndex+1; i++ {
		if len(p.list) > i+1 {
			dev := p.list[i+1].Key
			if len(dev) == 2 || (len(dev) == 5 && dev[2] == '-') {
				// probably langage tag (en-us etc..), ignore and continue loop
				continue
			}
			switch dev {
			case Chrome, Firefox, Safari, "Opera Mini", "Presto", "Version", "Mobile", "Mobile Safari", "Mozilla", "AppleWebKit", "Windows NT", "Windows Phone OS", Android, "Macintosh", Linux, "CrOS":
				// ignore this tokens, not device names
			default:
				if strings.Contains(strings.ToLower(dev), "tablet") {
					p.list[i+1].Key = "Tablet" // leave Tablet tag for later table detection
				} else {
					p.list = append(p.list[:i+1], p.list[i+2:]...)
				}
				return strings.TrimSpace(strings.TrimSuffix(dev, "Build"))
			}
		}
	}
	return ""
}
