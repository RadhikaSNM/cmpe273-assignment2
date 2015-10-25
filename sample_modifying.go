/*
Radhika SNM
009426196
*/
package main

import (
"fmt"
"gopkg.in/mgo.v2"
"gopkg.in/mgo.v2/bson"
"regexp"
"strings"
"github.com/julienschmidt/httprouter"
"net/http"
"encoding/json"
"errors"
"io/ioutil"
"net/url"
)

//var dbURL string ="mongodb://localhost:27017"
var dbURL string ="mongodb://DBTestUser:qwerty@ds048368.mongolab.com:48368/cmpe273db"
var ErrNotFound string= "not found"
var locationDBName string="cmpe273db"
var locationCollectionName string="locations"

//location: insert into db struct
type LocationInsert struct {
 Name string `json:"name"`
 Address string `json:"address"`
 City string `json:"city"`
 State string `json:"state"`
 Zip string `json:"zip"`

}

//location: db respose struct
type LocationDBResponse struct {
    Id bson.ObjectId `json:"id" bson:"_id,omitempty"`
    Name string `json:"name"`
    Address string `json:"address"`
    City string `json:"city"`
    State string `json:"state"`
    Zip string`json:"zip"`
    Coordinate Coordinates `json:"coordinate"`

    
}

//Coordinate struct
type Coordinates struct{
    Lat float64 `json:"lat"`
    Lng float64 `json:"lng"`
} 

//Response to be sent to the user - struct
type LocationInsertResponse struct {
    Id string `json:"id"`
    Name string `json:"name"`
    Address string `json:"address"`
    City string `json:"city"`
    State string `json:"state"`
    Zip string `json:"zip"`
    Coordinate Coordinates `json:"coordinate"`

}

//Google api result struct
type googleLocationResults struct{
    Results []struct{
        Geometry struct{
           Location struct{
            Lat float64 `json: lat`
            Lng float64  `json: lng`         
            }  `json: location`

            }  `json: geometry`
            } `json: results`
        }


//Error json struct
        type errorResponse struct{
            ErrorMessage string `json: errorMessage`
        }



        func createLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params){
            var errDecode error
            locationDetail:=LocationInsert{}

    //decode the sent json
            errDecode=json.NewDecoder(req.Body).Decode(&locationDetail)
            if errDecode!=nil{
                fmt.Println(errDecode.Error())
                msg:="Json sent was Empty/Incorrect .Error: "
                errorCheck(msg,rw)
                return
            }

            name:=locationDetail.Name
            address:=locationDetail.Address
            city:=locationDetail.City
            state:=locationDetail.State
            zip:=locationDetail.Zip
            fmt.Println("Name is :"+name,"Address: "+address)

   //Check if any of the expected fields are empty
            if(name==""||address==""||city==""||state==""||zip==""){
             msg1:="One or more of the fields in the sent json is missing "
             errorCheck(msg1,rw)
             return
         }


    //************************************
    //Calling the google api

         fullAdd:=address+","+city+","+state+","+zip
    //lat,long,errAdd:= getLatLong("1600 Amphitheatre Parkway Mountain View CA")
         lat,long,errAdd:= getLatLong(fullAdd)       

         if errAdd!=nil{
            errorCheck(errAdd.Error(),rw) 
            return

        }else{
            fmt.Println(lat)
            fmt.Println(long) 
        }

    //*******************************************
    //calling mongo lab

        session,c, err := connectToDB(dbURL,locationDBName,locationCollectionName)
        if err!=nil{
            errorCheck("Database connection error.",rw)
            return}
            defer session.Close()

            i := bson.NewObjectId()
            fmt.Println(i)
            fmt.Println("String version")
            idString:=i.String()
            fmt.Println(idString)

       //Extracting the ID
            r, err := regexp.Compile(`"[a-z0-9]+"`)
            split1:=r.FindString(idString)
            ID:=strings.Trim(split1,"\"")
            fmt.Println(ID)

            coord:=Coordinates{lat,long}
    //Inserting into the db
            d:=LocationDBResponse{i,name,address,city,state,zip,coord}
            err=c.Insert(d);

            if err != nil {
    //log.Fatal(err)
               fmt.Println("Database insertion error: ",err.Error())
               errorCheck("Database insertion error.",rw)
               return

           }
    //creation of the json response
           respStruct:=LocationInsertResponse{ID,name,address,city,state,zip,coord}
    //marshalling into a json

           respJson, err4 := json.Marshal(respStruct)
           if err4!=nil{
            fmt.Print("Error occcured in marshalling")
        }

    //sending it in the response
        rw.Header().Set("Content-Type","application/json")
        rw.WriteHeader(http.StatusCreated)
        fmt.Fprintf(rw, "%s", respJson)

    }




    func getLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

        idString:=p.ByName("location_id")

    //Check if given id is a valid hexademical string
        err_Hex:=checkHexString(idString)
        if err_Hex!=nil{
            errorCheck("The given location is not in the correct format",rw)
            return
        }

    //Obtaining the json from DB

    //MongoLab connection
        session,c, err := connectToDB(dbURL,locationDBName,locationCollectionName)
        if err!=nil{
            errorCheck("Database connection error.",rw)
            return}
            defer session.Close()


            result := LocationDBResponse{}
            err2 := c.Find(bson.M{"_id": bson.ObjectIdHex(idString)}).One(&result)
            fmt.Println(result)
            if err2 != nil {
        //log.Fatal(err2)
                errMsg:=err2.Error()
                fmt.Println("inside get- error")
                if err2.Error()==ErrNotFound{
                    errMsg="The given location id is incorrect. Please verify."
                }    
                errorCheck(errMsg,rw)
                return
            }

            fmt.Println("Name:", result.Name)
            fmt.Println("address",result.Address)
            fmt.Println("id",result.Id)


    //Marshling values into json
     //creating the json response
            respStruct:=LocationInsertResponse{idString,result.Name,result.Address,result.City,result.State,result.Zip,result.Coordinate}
            respJson, err4 := json.Marshal(respStruct)
            if err4!=nil{
                fmt.Print("Error occcured in marshalling")     
            }

    //sending it in the response
            rw.Header().Set("Content-Type","application/json")
            rw.WriteHeader(http.StatusOK)
            fmt.Fprintf(rw, "%s", respJson)

        }



        func updateLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params){

            locationDetail:=LocationInsert{}

    //decode the sent json
            errDecode:=json.NewDecoder(req.Body).Decode(&locationDetail)
            if errDecode!=nil{
                fmt.Println(errDecode.Error())
                msg:="Json sent was Empty/Incorrect"
                errorCheck(msg,rw)
                return
            }

            address:=locationDetail.Address
            city:=locationDetail.City
            state:=locationDetail.State
            zip:=locationDetail.Zip

            fmt.Println("Address: "+address,"city "+city,"state "+state)


        //Check if any of the expected fields are empty
            if(address==""||city==""||state==""||zip==""){
             msg1:="One or more of the fields in the sent json is missing "
             errorCheck(msg1,rw)
             return
         }



    //Obtain the id to modify:
         idString:=p.ByName("location_id")

        //Check if given id is a valid hexademical string
         err_Hex:=checkHexString(idString)
         if err_Hex!=nil{
            errorCheck("The given location is not in the correct format",rw)
            return
        }

    //************************************
    //Calling the google api

        fullAdd:=address+","+city+","+state+","+zip
    //lat,long,errAdd:= getLatLong("1600 Amphitheatre Parkway Mountain View CA")
        lat,long,errAdd:= getLatLong(fullAdd)       

        if errAdd!=nil{
            errorCheck(errAdd.Error(),rw) 
            return
        //TODO work about returning an empty
        }else{
            fmt.Println(lat)
            fmt.Println(long) 
        }

    //*******************************************

//MongoLab connection
        session,c, err := connectToDB(dbURL,locationDBName,locationCollectionName)
        if err!=nil{
            errorCheck("Database connection error.",rw)
            return}
            defer session.Close()


    //Modifying the data
            id:=bson.ObjectIdHex(idString)
            errUpdate := c.UpdateId(id,bson.M{"$set": bson.M{"address": address,"city":city,"state":state,"zip":zip,"coordinate.lat":lat,"coordinate.lng":long}})

            if errUpdate!=nil{
                fmt.Println("Inside update error")
                errMsg:=errUpdate.Error()
                if errUpdate.Error()==ErrNotFound{
                    fmt.Println("Inside ErrNotFound")
                    errMsg="The given location id is incorrect. Please verify."}
                    errorCheck(errMsg,rw)
                    return
                }

    //try to find the document again for the name:
                result := LocationDBResponse{}
                err2 := c.Find(bson.M{"_id": id}).One(&result)
                fmt.Println(result)
                if err2 != nil {
                    errorCheck(err2.Error(),rw)
                }

                coord:=Coordinates{lat,long}

    //send the json response
                respStruct:=LocationInsertResponse{idString,result.Name,address,city,state,zip,coord}
    //marshalling into a json

                respJson, err4 := json.Marshal(respStruct)
                if err4!=nil{
                    fmt.Print("Error occcured in marshalling")
                }

    //sending it in the response
                rw.Header().Set("Content-Type","application/json")
                rw.WriteHeader(http.StatusCreated)
                fmt.Fprintf(rw, "%s", respJson)

            }






            func deleteLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params){

                idString:=p.ByName("location_id")


         //Check if given id is a valid hexademical string
                err_Hex:=checkHexString(idString)
                if err_Hex!=nil{
                    errorCheck("The given location is not in the correct format",rw)
                    return
                }


                id:=bson.ObjectIdHex(idString)

    //*******************************************

    //MongoLab connection
                session,c, err := connectToDB(dbURL,locationDBName,locationCollectionName)
                if err!=nil{
                    errorCheck("Database connection error.",rw)
                    return}
                    defer session.Close()



    //Delete the id:
                    errDel:=c.RemoveId(id)
                    if errDel!=nil{
                     errMsg:=errDel.Error()
                     if errDel.Error()==ErrNotFound{
                        errMsg="The given location id is incorrect. Please verify."}
                        errorCheck(errMsg,rw)
                        return
                    }

    //Set the response
                    rw.WriteHeader(http.StatusOK)
    //fmt.Fprintf(rw, "%s", respJson)

                }


                func main() {
                    fmt.Println("=========================")
                    mux := httprouter.New()
                    mux.POST("/locations",createLocation)
                    mux.GET("/locations/:location_id", getLocation)
                    mux.PUT("/locations/:location_id", updateLocation)
                    mux.DELETE("/locations/:location_id", deleteLocation)


                    server := http.Server{
                        Addr:        "0.0.0.0:8080",
                        Handler: mux,
                    }

                    server.ListenAndServe()

                }


 //Send error json          
                func errorCheck(errMsg string,rw http.ResponseWriter){
    //Creating a errorJson
                    errorSt:=errorResponse{errMsg}

                    errorJson, err4 := json.Marshal(errorSt)
                    if err4!=nil{
                        fmt.Print("Error occcured in marshalling")
                    }
                    rw.Header().Set("Content-Type","application/json")
                    rw.WriteHeader(http.StatusBadRequest)
                    fmt.Fprintf(rw, "%s", errorJson)


                }

//Get lat longitude from google api
                func getLatLong(CombinedAddress string) (float64,float64,error){
                    addressEnc:=url.QueryEscape(CombinedAddress)
                    fmt.Println("combined link"+addressEnc)
                    link:="http://maps.google.com/maps/api/geocode/json?address="+addressEnc
                    fmt.Println(link)
                    resp, err := http.Get(link);
                    if err != nil {
                        fmt.Println(err.Error())
                        err_google:=errors.New("Google map api connection could not be established!")
                        return 0,0,err_google
                    }

                    defer resp.Body.Close()
                    body, err1 := ioutil.ReadAll(resp.Body)
                    if err1 != nil {
                        return 0,0,err1

                    }
                    var location googleLocationResults
                //Unmarshall the response into a json
                    err2:=json.Unmarshal(body,&location)
                    if err2 != nil {
                        return 0,0,err2
                    }


    //check is results is null
                    if (len(location.Results)==0){
                        err_noRes:=errors.New("The provided address is invalid. Please Check")
                        return 0,0,err_noRes 

                    }


                    lat:=location.Results[0].Geometry.Location.Lat
                    lng:=location.Results[0].Geometry.Location.Lng

                    return lat,lng,nil
                }


//Function to connect to the database
                func connectToDB(dbURL string, dbName string, collectionName string) (*mgo.Session,*mgo.Collection,error){
                  session, err := mgo.Dial(dbURL)
                  if err != nil {
                    fmt.Println("Database connection error: ",err.Error())
                    return nil,nil,err

                }
        // Optional. Switch the session to a monotonic behavior.
                session.SetMode(mgo.Monotonic, true)
                c := session.DB(dbName).C(collectionName)  
                return session,c,nil
            }

            func checkHexString(id string) error{
             stringFormat,err:=regexp.MatchString("^[A-Fa-f0-9]{24}$", id)
             if err!=nil{
                return err
            }else if(!stringFormat){
                err_Format:=errors.New("Given location id is not in a valid format.")
                return err_Format 
            }else{
                return nil
            }


        }



