package controller

import "errors"

var (
	errPkgInfoNameIsDuplicated            = errors.New("the package info name must be unique in the same registry")
	errInvalidType                        = errors.New("type is invalid")
	errConfigFileNotFound                 = errors.New("configuration file isn't found")
	errUnknownPkg                         = errors.New("unknown package")
	errGitHubReleaseTypeAssertion         = errors.New("pkg typs is github_release, but it isn't *GitHubReleasePackageInfo")
	errTypeAssertionHTTPPackageInfo       = errors.New("pkg typs is http, but it isn't *HTTPPackageInfo")
	errTypeAssertionGitHubContentRegistry = errors.New("registry.GetType() is github_content, but registry isn't *GitHubContentRegistry")
	errInvalidPackageType                 = errors.New("package type is invalid")
	errGitHubTokenIsRequired              = errors.New("GITHUB_TOKEN is required for the type `github_release`")
	errCommandIsRequired                  = errors.New("command is required")
	errCommandIsNotFound                  = errors.New("command is not found")
	errGitHubContentMustBeFile            = errors.New("ref must be not a directory but a file")
	errUnsupportedRegistryType            = errors.New("unsupported registry type")
	errLocalRegistryNotFound              = errors.New("local registry isn't found")
	errRegistryNotFound                   = errors.New("registry isn't found")
	errPkgNotFound                        = errors.New("package isn't found in the registry")
	errExePathIsDirectory                 = errors.New("exe_path is directory")
	errChmod                              = errors.New("add the permission to execute the command")
	errInvalidHTTPStatusCode              = errors.New("status code >= 400")
	errInstallFailure                     = errors.New("it failed to install some packages")
)
