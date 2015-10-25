package main

import (
"encoding/json"
"fmt"
"net/http"
"github.com/julienschmidt/httprouter"
"gopkg.in/mgo.v2"
"strings"
"gopkg.in/mgo.v2/bson"
"io/ioutil"
)



type Startresults struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string   `json:"place_id"`
		Types   []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

	type postresp struct {

      Id bson.ObjectId `json:"id" bson:"_id"`
      Name string `json:"name" bson:"name"`
      Address string `json:"address" bson:"address"`
      City string `json:"city" bson:"city"`
      State string `json:"state" bson:"state"`
      Zip string `json:"zip" bson:"zip"`
			Loc Cord `json:"coordinate" bson:"coordinate"`
    }

		type Cord struct {
						Lat float64 `json:"lat" bson:"lat"`
						Lng float64 `json:"lng" bson:"lng"`
		}

		type LocNavigator struct {
				session *mgo.Session
			}

func NewNavigator(s *mgo.Session) *LocNavigator {
	return &LocNavigator{s}
}

func (ln LocNavigator) GetLoc(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")

	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	oid := bson.ObjectIdHex(id)

	po := postresp{}


	if err := ln.session.DB("ojas").C("locationsC").FindId(oid).One(&po); err != nil {
		w.WriteHeader(404)
		return
	}

	json.NewDecoder(r.Body).Decode(po)

	uj, _ := json.Marshal(po)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", uj)
}

func (ln LocNavigator) UpdateLoc (w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	id := p.ByName("id")

	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	oid := bson.ObjectIdHex(id)

	po := postresp{}
	ps := postresp{}

	ps.Id = oid

	json.NewDecoder(r.Body).Decode(&ps)

	if err := ln.session.DB("ojas").C("locationsC").FindId(oid).One(&po); err != nil {
		w.WriteHeader(404)
		return
	}

	na := po.Name

	collections := ln.session.DB("ojas").C("locationsC")

	po = fetchdata(&ps)
	collections.Update(bson.M{"_id": oid}, bson.M{"$set": bson.M{ "address": ps.Address, "city": ps.City, "state": ps.State, "zip" : ps.Zip, "coordinate": bson.M{"lat" : po.Loc.Lat, "lng" : po.Loc.Lng}}})

	po.Name = na

	uj, _ := json.Marshal(po)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)

}

func (ln LocNavigator) RemoveLoc(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	    id := p.ByName("id")

	    if !bson.IsObjectIdHex(id) {
	        w.WriteHeader(404)
	        return
	    }

	    oid := bson.ObjectIdHex(id)

	    if err := ln.session.DB("ojas").C("locationsC").RemoveId(oid); err != nil {
	        w.WriteHeader(404)
	        return
	    }
	    w.WriteHeader(200)
}


func (ln LocNavigator) CreateLoc(w http.ResponseWriter, r *http.Request, p httprouter.Params) {


	postrs := postresp{}

	json.NewDecoder(r.Body).Decode(&postrs)

	neww := fetchdata(&postrs)

	neww.Id = bson.NewObjectId()


	ln.session.DB("ojas").C("locationsC").Insert(neww)


	uj, _ := json.Marshal(neww)


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)
}

func fetchdata (rep *postresp) postresp {

  add := rep.Address

  ci := rep.City

  gs := strings.Replace(rep.State," ","+",-1)
  gadd := strings.Replace(add, " ", "+", -1)
  gci := strings.Replace(ci," ","+",-1)

	uri := "http://maps.google.com/maps/api/geocode/json?address="+gadd+"+"+gci+"+"+gs+"&sensor=false"


  resp, _ := http.Get(uri)

	body, _ := ioutil.ReadAll(resp.Body)


 	C := Startresults{}

  err := json.Unmarshal(body, &C)
   if err!= nil {
     panic(err)
   }


	 for _, Sample := range C.Results {
				rep.Loc.Lat= Sample.Geometry.Location.Lat
				rep.Loc.Lng = Sample.Geometry.Location.Lng
		}

   return *rep
}


func getSession() *mgo.Session {



	s, err := mgo.Dial("mongodb://localhost:27017")


    if err != nil {
        panic(err)
    }
    return s
}

func main() {

  		r := httprouter.New()

  	ln := NewNavigator(getSession())


  	r.GET("/locations/:id", ln.GetLoc)
  	r.POST("/locations",ln.CreateLoc)
		r.PUT("/locations/:id",ln.UpdateLoc)
		r.DELETE("/locations/:id",ln.RemoveLoc)

		http.ListenAndServe("localhost:3025",r)

}




