package main

import (
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getFileExtension (mediaType string) (extension string, err error) {
	parsedType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return "", err
	}	

	extensions, err := mime.ExtensionsByType(parsedType)
	if err != nil {
		return "", err
	} else if len(extensions) < 1 {
		return "", errors.New("invalid extension type for thumbnail")
	}
	return extensions[0], nil 
}




func (cfg *apiConfig) createFolderAndFile (videoID, ext string) (thumbFile *os.File, err error) {
	err = os.MkdirAll(cfg.assetsRoot, 0755)
    if err != nil {
        return nil, err
    }
	
	thumbFilePath := filepath.Join(cfg.assetsRoot, videoID + ext)
	thumbFile, err = os.Create(thumbFilePath)
	if err != nil {
		return nil, err
	}
	// i dont think the following goes here? probably goes right after the function call to this fucntion
	// defer thumbFile.Close()

	return thumbFile, nil

}

func (cfg *apiConfig) getAssetURL(assetPath string) string {
	url := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
	return url
}
	