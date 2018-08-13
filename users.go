package bitbucket

type Users struct {
	c *Client
}

func (u *Users) Get(t string) (interface{}, error) {

	urlStr := GetApiBaseURL() + "/users/" + t + "/"
	return u.c.execute("GET", urlStr, "")
}

func (c *Client) Get(t string) (interface{}, error) {

	urlStr := GetApiBaseURL() + "/users/" + t + "/"
	return c.execute("GET", urlStr, "")
}

func (u *Users) Followers(t string) (interface{}, error) {

	urlStr := GetApiBaseURL() + "/users/" + t + "/followers"
	return u.c.execute("GET", urlStr, "")
}

func (u *Users) Following(t string) (interface{}, error) {

	urlStr := GetApiBaseURL() + "/users/" + t + "/following"
	return u.c.execute("GET", urlStr, "")
}
func (u *Users) Repositories(t string) (interface{}, error) {

	urlStr := GetApiBaseURL() + "/users/" + t + "/repositories"
	return u.c.execute("GET", urlStr, "")
}
/*func (c *Client) Projects() ([]interface{}) {

	var allProjects []interface{}
	//List his teams
	res, err := c.Teams.List("member")
	if err != nil{
		panic(err)
	}

	var teams []string
    datas := res.(map[string]interface{})["values"].([]interface{})
    for i := 0; i < len(datas); i++{
        teams = append(teams, datas[i].(map[string]interface{})["username"].(string))
    }
	//For each teams get projects
    for i := 0; i < len(teams); i++{
    	projects, err2 := c.Teams.Projects(teams[i])
    	if err2 != nil{
			panic(err2)
		}

		allProjects = append(allProjects, projects)
    }
    return allProjects

}*/