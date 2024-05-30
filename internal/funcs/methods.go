package funcs

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
	"github.com/SrgBr-0/TaskPizSuWok/internal/entity"
)

const registryURL = "http://localhost:5000/v2"

var containerCache = cache.New(5*time.Minute, 10*time.Minute)
var ipRequestTimes = make(map[string]time.Time)
var mu sync.Mutex

func GetContainerInfoHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	ip := r.Header.Get("X-Forwarded-For")
	mu.Lock()
	if lastRequestTime, exists := ipRequestTimes[ip]; exists {
		if time.Since(lastRequestTime) < time.Second {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			mu.Unlock()
			return
		}
	}
	ipRequestTimes[ip] = time.Now()
	mu.Unlock()

	if data, found := containerCache.Get(name); found {
		log.Printf("Cache hit for %s", name)
		json.NewEncoder(w).Encode(data)
		return
	}
	log.Printf("Cache miss for %s", name)

	tags, err := getTags(name)
	if err != nil {
		log.Printf("Failed to get tags for %s: %v", name, err)
		http.Error(w, "Failed to get tags", http.StatusInternalServerError)
		return
	}

	infoMap := make(map[string]entity.ContainerInfo)
	for _, tag := range tags {
		info, err := getManifestInfo(name, tag)
		if err != nil {
			log.Printf("Failed to get manifest info for %s:%s: %v", name, tag, err)
			continue
		}
		infoMap[fmt.Sprintf("%s:%s", name, tag)] = info
	}

	containerCache.Set(name, infoMap, cache.DefaultExpiration)
	json.NewEncoder(w).Encode(infoMap)
}

func getTags(name string) ([]string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s/tags/list", registryURL, name))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error response from registry for tags list: %s", body)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Tags, nil
}

func getManifestInfo(name, tag string) (entity.ContainerInfo, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/manifests/%s", registryURL, name, tag), nil)
	if err != nil {
		return entity.ContainerInfo{}, err
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return entity.ContainerInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error response from registry for manifest: %s", body)
		return entity.ContainerInfo{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var manifest struct {
		Layers []struct {
			Size int `json:"size"`
		} `json:"layers"`
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return entity.ContainerInfo{}, err
	}
	log.Printf("Manifest response: %s", body)

	if err := json.Unmarshal(body, &manifest); err != nil {
		log.Printf("Failed to decode manifest response: %s", body)
		return entity.ContainerInfo{}, err
	}

	var totalSize int
	for _, layer := range manifest.Layers {
		totalSize += layer.Size
	}

	return entity.ContainerInfo{
		Layers: len(manifest.Layers),
		Size:   totalSize,
	}, nil
}
