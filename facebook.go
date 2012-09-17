package facebook

import (
	"net/http"
	"fmt"
	"strconv"
	"strings"
	"encoding/json"
	"github.com/chaosphere2112/go-jsonUtil"	
)

type AccessToken struct {
	Token string
	Expiry int64
}

func readHttpBody(response *http.Response) string {

	bodyBuffer := make([]byte, 1000)
	var str string
	
	count, err := response.Body.Read(bodyBuffer)

	for ; count > 0; count, err = response.Body.Read(bodyBuffer) {

		if err != nil {

		}

		str += string(bodyBuffer[:count])
	}

	return str

}

//Converts a code to an Auth_Token
func GetAccessToken(client_id string, code string, secret string, callbackUri string) AccessToken {
	
	//https://graph.facebook.com/oauth/access_token?client_id=YOUR_APP_ID&redirect_uri=YOUR_REDIRECT_URI&client_secret=YOUR_APP_SECRET&code=CODE_GENERATED_BY_FACEBOOK
	response, err := http.Get("https://graph.facebook.com/oauth/access_token?client_id="+
		client_id+"&redirect_uri="+callbackUri+
		"&client_secret="+secret+"&code="+code)

	if err == nil {
		
		auth := readHttpBody(response)

	    
	    var token AccessToken

	    tokenArr := strings.Split(auth, "&")

	    token.Token = strings.Split(tokenArr[0], "=")[1]
	    expireInt,err := strconv.Atoi(strings.Split(tokenArr[1],"=")[1])

	    if (err == nil) {
		    token.Expiry = int64(expireInt)
		}

	    return token
	}

	var token AccessToken

	return token
}

func GetMe(token AccessToken) string {

	response, err := getUncachedResponse("https://graph.facebook.com/me?access_token="+token.Token)

	if err == nil {
		
		var jsonBlob interface{}

		responseBody := readHttpBody(response)

		if responseBody != "" {
			err = json.Unmarshal([]byte(responseBody), &jsonBlob)

			if err == nil {
				jsonObj := jsonBlob.(map[string]interface{})
				return jsonObj["id"].(string)
			}
		}
		return err.Error()
	}

	return err.Error()
}

func getUncachedResponse(uri string) (*http.Response, error) {

	request, err := http.NewRequest("GET", uri, nil)

	if err == nil {
		request.Header.Add("Cache-Control", "no-cache")

		client := new(http.Client)

		return client.Do(request)
	}

	if (err != nil) {
	}
	return nil, err

}

func getPhotoSource(token *AccessToken, photoId string) string {
	
	response, err := getUncachedResponse("https://graph.facebook.com/"+photoId+"?access_token="+token.Token+"&fields=source")

	if err == nil && response != nil {

		body := readHttpBody(response)

		if body != "" {

			object, err := jsonUtil.JsonFromString(body)

			if err == nil {

				source, err := object.String("source")

				if err == nil {

					return source

				}

			}

		}

	}

	return ""

}

func GetAlbumPhotos(token *AccessToken, albumId string)  []string {
	response, err := getUncachedResponse("https://graph.facebook.com/"+albumId+"/photos?access_token="+token.Token+"&fields=images&limit=1000")

	if err == nil && response != nil {

		body := readHttpBody(response)

		if body != "" {

			object,err := jsonUtil.JsonFromString(body)
			if err == nil {
				//Get the "data" array
				var data jsonUtil.JsonArray
				returned := make([]string, 0, 1)

				data, err = object.Array("data")

				if err == nil {
					fmt.Println(len(data))
					for dataIndex:=0; dataIndex < len(data); dataIndex++ {
						var anonObj jsonUtil.JsonObject
						anonObj, err = data.Object(dataIndex)
						if err == nil {

							//For each object in data, iterate over the "images" array
							var imagesArray jsonUtil.JsonArray

							imagesArray,err = anonObj.Array("images")
							if err == nil {
								//Get the source for each image
								var image jsonUtil.JsonObject
								image, err = imagesArray.Object(1)
								
								if err == nil {
									var photoSource string
									photoSource, err = image.String("source")

									if err == nil {
										returned = append(returned, photoSource)
									}
								} 
							}

						}

					}

				}
				return returned
			}
		}

	}

	if err != nil {
		returned := make([]string, 1)
		returned[0] = err.Error()
		return returned
	}

	return make([]string, 0)
}

func GetPhotos(token *AccessToken) []string {

	response, err := getUncachedResponse("https://graph.facebook.com/me/albums?access_token="+token.Token+"&fields=id")

	if err == nil && response != nil {
		var jsonBlob interface{}

		responseBody := readHttpBody(response)

		if responseBody != "" {
			err = json.Unmarshal([]byte(responseBody), &jsonBlob)

			if err == nil {
				jsonObj := jsonBlob.(map[string]interface{})
				
	
				dataArray := jsonObj["data"].([]interface{})

				first := dataArray[0].(map[string]interface{})

				firstAlbumId := first["id"].(string)

				//Feed albumId into GetAlbumPhotos
				return GetAlbumPhotos(token, firstAlbumId)

			}
		}
	}
	fmt.Println("Failed to GetPhotos because "+err.Error())
	return nil
}
