package api

import (
	"cf/configuration"
	"cf/net"
	"strings"
)

type EndpointRepository interface {
	UpdateEndpoint(endpoint string) (finalEndpoint string, apiResponse net.ApiResponse)
	GetLoggregatorEndpoint() (endpoint string, apiResponse net.ApiResponse)
	GetUAAEndpoint() (endpoint string, apiResponse net.ApiResponse)
	GetCloudControllerEndpoint() (endpoint string, apiResponse net.ApiResponse)
}

type RemoteEndpointRepository struct {
	config     *configuration.Configuration
	gateway    net.Gateway
	configRepo configuration.ConfigurationRepository
}

func NewEndpointRepository(config *configuration.Configuration, gateway net.Gateway, configRepo configuration.ConfigurationRepository) (repo RemoteEndpointRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.configRepo = configRepo
	return
}

func (repo RemoteEndpointRepository) UpdateEndpoint(endpoint string) (finalEndpoint string, apiResponse net.ApiResponse) {
	endpointMissingScheme := !strings.HasPrefix(endpoint, "https://") && !strings.HasPrefix(endpoint, "http://")

	if endpointMissingScheme {
		finalEndpoint = "https://" + endpoint
		apiResponse = repo.attemptUpdate(finalEndpoint)

		if apiResponse.IsNotSuccessful() {
			finalEndpoint = "http://" + endpoint
			apiResponse = repo.attemptUpdate(finalEndpoint)
		}
		return
	}

	finalEndpoint = endpoint

	apiResponse = repo.attemptUpdate(finalEndpoint)

	return
}

func (repo RemoteEndpointRepository) attemptUpdate(endpoint string) (apiResponse net.ApiResponse) {
	request, apiResponse := repo.gateway.NewRequest("GET", endpoint+"/v2/info", "", nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	serverResponse := new(struct {
		ApiVersion            string `json:"api_version"`
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		LoggregatorEndpoint   string `json:"logging_endpoint"`
	})
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, &serverResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if endpoint != repo.config.Target {
		repo.configRepo.ClearSession()
	}

	repo.config.Target = endpoint
	repo.config.ApiVersion = serverResponse.ApiVersion
	repo.config.AuthorizationEndpoint = serverResponse.AuthorizationEndpoint
	repo.config.LoggregatorEndPoint = serverResponse.LoggregatorEndpoint

	err := repo.configRepo.Save()
	if err != nil {
		apiResponse = net.NewApiResponseWithMessage(err.Error())
	}
	return
}

func (repo RemoteEndpointRepository) GetLoggregatorEndpoint() (endpoint string, apiResponse net.ApiResponse) {
	if repo.config.LoggregatorEndPoint == "" {
		apiResponse = net.NewApiResponseWithMessage("Loggregator endpoint missing from config file")
		return
	}

	endpoint = repo.config.LoggregatorEndPoint
	return
}

func (repo RemoteEndpointRepository) GetCloudControllerEndpoint() (endpoint string, apiResponse net.ApiResponse) {
	if repo.config.Target == "" {
		apiResponse = net.NewApiResponseWithMessage("Target endpoint missing from config file")
		return
	}

	endpoint = repo.config.Target
	return
}

func (repo RemoteEndpointRepository) GetUAAEndpoint() (endpoint string, apiResponse net.ApiResponse) {
	if repo.config.AuthorizationEndpoint == "" {
		apiResponse = net.NewApiResponseWithMessage("UAA endpoint missing from config file")
		return
	}

	endpoint = strings.Replace(repo.config.AuthorizationEndpoint, "login", "uaa", 1)

	return
}
