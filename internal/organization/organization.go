package organization

import (
	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	"github.com/moneyforward-i/admina-sysutils/internal/logger"
)

type Client interface {
	GetOrganization() (*admina.Organization, error)
}

// PrintInfo prints organization information in a formatted way
func PrintInfo(org *admina.Organization) {
	logger.LogInfo("-----------------------------------------------------------------")
	logger.LogInfo("%s | %s | %d (%s)", org.Name, org.UniqueName, org.ID, org.Status)
	logger.LogInfo("Language: %s | Location: %s | TimeZone: %s", org.SystemLanguage, org.Location, org.TimeZone)
	logger.LogInfo("Domains: %v", org.Domains)
	logger.LogInfo("-----------------------------------------------------------------")
}
