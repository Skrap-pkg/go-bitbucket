package bitbucket

import (
	"encoding/json"
	"fmt"
	"log"

	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	//"net/http/httputil"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/bitbucket"
	"golang.org/x/oauth2/clientcredentials"
)

const DEFAULT_PAGE_LENGTH = 10

type Client struct {
	Auth         *auth
	Users        users
	User         user
	Teams        teams
	Repositories *Repositories
	Pagelen      uint64
}

type auth struct {
	appID, secret  string
	user, password string
	token          oauth2.Token
	bearerToken    string
}

// Uses the Client Credentials Grant oauth2 flow to authenticate to Bitbucket
func NewOAuthClientCredentials(i, s string) *Client {
	a := &auth{appID: i, secret: s}
	ctx := context.Background()
	conf := &clientcredentials.Config{
		ClientID:     i,
		ClientSecret: s,
		TokenURL:     bitbucket.Endpoint.TokenURL,
	}

	tok, err := conf.Token(ctx)
	if err != nil {
		log.Fatal(err)
	}
	a.token = *tok
	return injectClient(a)

}

func NewOAuth(i, s string) *Client {
	a := &auth{appID: i, secret: s}
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     i,
		ClientSecret: s,
		Endpoint:     bitbucket.Endpoint,
	}

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog:\n%v", url)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	fmt.Printf("Enter the code in the return URL: ")
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}
	a.token = *tok
	return injectClient(a)
}

func NewOAuthbearerToken(t string) *Client {
	a := &auth{bearerToken: t}
	return injectClient(a)
}

func NewBasicAuth(u, p string) *Client {
	a := &auth{user: u, password: p}
	return injectClient(a)
}

func injectClient(a *auth) *Client {
	c := &Client{Auth: a, Pagelen: DEFAULT_PAGE_LENGTH}
	c.Repositories = &Repositories{
		c:                  c,
		PullRequests:       &PullRequests{c: c},
		Repository:         &Repository{c: c},
		Commits:            &Commits{c: c},
		Diff:               &Diff{c: c},
		BranchRestrictions: &BranchRestrictions{c: c},
		Webhooks:           &Webhooks{c: c},
	}
	c.Users = &Users{c: c}
	c.User = &User{c: c}
	c.Teams = &Teams{c: c}
	return c
}

func (c *Client) execute(method string, urlStr string, text string) (interface{}, error) {
	// Use pagination if changed from default value
	const DEC_RADIX = 10
	if strings.Contains(urlStr, "/repositories/") {
		if c.Pagelen != DEFAULT_PAGE_LENGTH {
			urlObj, err := url.Parse(urlStr)
			if err != nil {
				return nil, err
			}
			q := urlObj.Query()
			q.Set("pagelen", strconv.FormatUint(c.Pagelen, DEC_RADIX))
			urlObj.RawQuery = q.Encode()
			urlStr = urlObj.String()
		}
	}

	body := strings.NewReader(text)

	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	if text != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.Auth.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.Auth.bearerToken)
	}

	if c.Auth.user != "" && c.Auth.password != "" {
		req.SetBasicAuth(c.Auth.user, c.Auth.password)
	} else if c.Auth.token.Valid() {
		c.Auth.token.SetAuthHeader(req)
	}

	//DEBUG
	/*dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%q", dump)*/
	//

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if (resp.StatusCode != http.StatusOK) && (resp.StatusCode != http.StatusCreated) {
		return nil, fmt.Errorf(resp.Status)
	}

	if resp.Body == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	resBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result interface{}
	err = json.Unmarshal(resBodyBytes, &result)
	if err != nil {
		return nil, err
	}

	resultMap, isMap := result.(map[string]interface{})
	if isMap {
		nextIn := resultMap["next"]
		valuesIn := resultMap["values"]
		if nextIn != nil && valuesIn != nil {
			nextUrl := nextIn.(string)
			if nextUrl != "" {
				valuesSlice := valuesIn.([]interface{})
				if valuesSlice != nil {
					nextResult, err := c.execute(method, nextUrl, text)
					if err != nil {
						return nil, err
					}
					nextResultMap, isNextMap := nextResult.(map[string]interface{})
					if !isNextMap {
						return nil, fmt.Errorf("next page result is not map, it's %T", nextResult)
					}
					nextValuesIn := nextResultMap["values"]
					if nextValuesIn == nil {
						return nil, fmt.Errorf("next page result has no values")
					}
					nextValuesSlice, isSlice := nextValuesIn.([]interface{})
					if !isSlice {
						return nil, fmt.Errorf("next page result 'values' is not slice")
					}
					valuesSlice = append(valuesSlice, nextValuesSlice...)
					resultMap["values"] = valuesSlice
					delete(resultMap, "page")
					delete(resultMap, "pagelen")
					delete(resultMap, "size")
					result = resultMap
				}
			}
		}
	}

	return result, nil
}

func (c *Client) requestUrl(template string, args ...interface{}) string {

	if len(args) == 1 && args[0] == "" {
		return GetApiBaseURL() + template
	}
	return GetApiBaseURL() + fmt.Sprintf(template, args...)
}
