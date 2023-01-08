package repository

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type SubscriptionRepository interface {
	DeleteModFile(fileName string) error
	DownloadModFile(modID string, fileName string) error
	DoesFileExist(modID string) bool
}

type fileRepository struct {
	modFolder string
}

func NewRepository(modFolder string) *fileRepository {
	return &fileRepository{modFolder: modFolder}
}

func (p *fileRepository) DoesFileExist(modID string) bool {
	_, error := os.Stat(fmt.Sprintf(`%s\%s.pnt`, p.modFolder, modID))

	if os.IsNotExist(error) {
		return false
	} else {
		return true
	}
}

func (p *fileRepository) DeleteModFile(fileName string) error {
	err := os.Remove(fmt.Sprintf(`%s\%s.pnt`, p.modFolder, fileName))
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func (p *fileRepository) DownloadModFile(modID string, fileName string) error {
	// Create the file
	out, err := os.Create(fmt.Sprintf(`%s\%s.pnt`, p.modFolder, fileName))
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(fmt.Sprintf(`http://localhost:9000/mods/%s.pnt`, modID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		out.Close()
		err := os.Remove(fmt.Sprintf(`%s\%s.pnt`, p.modFolder, fileName))
		if err != nil {
			fmt.Println(err)
		}

		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
