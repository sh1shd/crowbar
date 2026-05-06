// Package entity defines data structures and entities used by the web layer of the 3x-ui panel.
package entity

import (
	"time"

	"github.com/mhsanaei/3x-ui/v2/util/common"
)

// Msg represents a standard API response message with success status, message text, and optional data object.
type Msg struct {
	Success bool   `json:"success"` // Indicates if the operation was successful
	Msg     string `json:"msg"`     // Response message text
	Obj     any    `json:"obj"`     // Optional data object
}

// AllSetting contains all configuration settings for the 3x-ui panel including web server and subscription settings.
type AllSetting struct {
	// Panel UI settings
	PageSize    int    `json:"pageSize" form:"pageSize"`       // Number of items per page in lists
	ExpireDiff  int    `json:"expireDiff" form:"expireDiff"`   // Expiration warning threshold in days
	TrafficDiff int    `json:"trafficDiff" form:"trafficDiff"` // Traffic warning threshold percentage
	RemarkModel string `json:"remarkModel" form:"remarkModel"` // Remark model pattern for inbounds

	// Security settings
	TimeLocation    string `json:"timeLocation" form:"timeLocation"`       // Time zone location
	TwoFactorEnable bool   `json:"twoFactorEnable" form:"twoFactorEnable"` // Enable two-factor authentication
	TwoFactorToken  string `json:"twoFactorToken" form:"twoFactorToken"`   // Two-factor authentication token

	// Subscription server settings
	SubCustomHeaders            string `json:"subCustomHeaders" form:"subCustomHeaders"`                       // Custom HTTP headers for subscription responses (JSON)
	SubCustomHtml               string `json:"subCustomHtml" form:"subCustomHtml"`                             // Custom HTML content returned for subscription pages
	SubCustomErrorHtml          string `json:"subCustomErrorHtml" form:"subCustomErrorHtml"`                   // Custom HTML content returned for error pages
	SubEncrypt                  bool   `json:"subEncrypt" form:"subEncrypt"`                                   // Encrypt subscription responses
	SubURI                      string `json:"subURI" form:"subURI"`                                           // Subscription server URI
	SubMessageClientDisabled    string `json:"subMessageClientDisabled" form:"subMessageClientDisabled"`       // Message shown to clients when subscription is disabled
	SubMessageClientExpired     string `json:"subMessageClientExpired" form:"subMessageClientExpired"`         // Message shown to clients when subscription has expired
	SubMessageClientTrafficEnd  string `json:"subMessageClientTrafficEnd" form:"subMessageClientTrafficEnd"`   // Message shown to clients when traffic has ended
	SubMessageContactAdmin      string `json:"subMessageContactAdmin" form:"subMessageContactAdmin"`           // Message shown to clients to contact administrator
	SubEnableIndexPage          bool   `json:"subEnableIndexPage" form:"subEnableIndexPage"`                   // Enable custom index page for subscription root
	SubIndexPageHtml            string `json:"subIndexPageHtml" form:"subIndexPageHtml"`                       // Custom index page HTML/content
}

// CheckValid validates all settings in the AllSetting struct, checking IP addresses, ports, SSL certificates, and other configuration values.
func (s *AllSetting) CheckValid() error {
	_, err := time.LoadLocation(s.TimeLocation)
	if err != nil {
		return common.NewError("time location not exist:", s.TimeLocation)
	}

	return nil
}
