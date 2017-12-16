package commands

import "errors"

var (
	ErrorStoreMissingFunctionName    = errors.New("please provide the function name")
	ErrorMissingFunctionName         = errors.New("please provide a name for the function")
	ErrorMissingFunctionNameFlag     = errors.New("please provide the function name with --name")
	ErrorMissingImageFlag            = errors.New("please provide the image name with --image")
	ErrorMissingUsername             = errors.New("must provide --username or -u")
	ErrorPasswordStdinAndPassword    = errors.New("--password and --password-stdin are mutually exclusive")
	ErrorMissingPassword             = errors.New("must provide a non-empty password via --password or --password-stdin")
	ErrorMissingGateway              = errors.New("gateway cannot be an empty string")
	ErrorUnauthorizedGateway         = errors.New("unauthorized access, run \"faas-cli login\" to setup authentication for this server")
	ErrorInvalidYamlFile             = errors.New("you must supply a valid YAML file")
	ErrorMissingLangFlagFile         = errors.New("you must supply a function language with the --lang flag")
	ErrorInvalidImageFlag            = errors.New("please provide a valid image name with --image for your Docker image")
	ErrorFunctionPath                = errors.New("please provide the full path to your function's handler")
	ErrorMissingFunctionNameToDelete = errors.New("please provide the name of a function to delete")
	ErrorUnavailableLanguageTemplate = errors.New("no language templates were found. Please run 'faas-cli template pull'")
	ErrorExclusiveUpdateReplaceFlag  = errors.New("cannot specify --update and --replace at the same time")
	ErrorGatewayUnsuccessfulLogin    = errors.New("unable to login, either username or password is incorrect")
	ErrorInvalidLabel                = errors.New("label format is not correct, needs key=value")
	ErrorBashCompletionFilename      = errors.New("please provide filename for bash completion")
	ErrorBashCompletionFileUncreated = errors.New("unable to create bash completion file")
	ErrorInvalidRepositoryURL        = errors.New("the repository URL must be a valid git repo uri")
)
