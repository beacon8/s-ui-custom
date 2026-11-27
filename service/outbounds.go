package service

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/util/common"

	"gorm.io/gorm"
)

type OutboundService struct{}

func (o *OutboundService) GetAll() (*[]map[string]interface{}, error) {
	db := database.GetDB()
	outbounds := []*model.Outbound{}
	err := db.Model(model.Outbound{}).Scan(&outbounds).Error
	if err != nil {
		return nil, err
	}
	var data []map[string]interface{}
	for _, outbound := range outbounds {
		outData := map[string]interface{}{
			"id":   outbound.Id,
			"type": outbound.Type,
			"tag":  outbound.Tag,
		}
		if outbound.Options != nil {
			var restFields map[string]json.RawMessage
			if err := json.Unmarshal(outbound.Options, &restFields); err != nil {
				return nil, err
			}
			for k, v := range restFields {
				outData[k] = v
			}
		}
		data = append(data, outData)
	}
	return &data, nil
}

func (o *OutboundService) GetAllConfig(db *gorm.DB) ([]json.RawMessage, error) {
	var outboundsJson []json.RawMessage
	var outbounds []*model.Outbound
	err := db.Model(model.Outbound{}).Scan(&outbounds).Error
	if err != nil {
		return nil, err
	}
	for _, outbound := range outbounds {
		outboundJson, err := outbound.MarshalJSON()
		if err != nil {
			return nil, err
		}
		outboundsJson = append(outboundsJson, outboundJson)
	}
	return outboundsJson, nil
}

func (s *OutboundService) Save(tx *gorm.DB, act string, data json.RawMessage) error {
	var err error

	switch act {
	case "new", "edit":
		var outbound model.Outbound
		err = outbound.UnmarshalJSON(data)
		if err != nil {
			return err
		}

		var oldTag string
		if act == "edit" {
			err = tx.Model(model.Outbound{}).Select("tag").Where("id = ?", outbound.Id).Find(&oldTag).Error
			if err != nil {
				return err
			}
		}

		if corePtr.IsRunning() {
			configData, err := outbound.MarshalJSON()
			if err != nil {
				return err
			}
			if act == "edit" {
				err = corePtr.RemoveOutbound(oldTag)
				if err != nil && err != os.ErrInvalid {
					return err
				}
			}
			err = corePtr.AddOutbound(configData)
			if err != nil {
				return err
			}
		}

		err = tx.Save(&outbound).Error
		if err != nil {
			return err
		}
	case "del":
		var tag string
		err = json.Unmarshal(data, &tag)
		if err != nil {
			return err
		}
		if corePtr.IsRunning() {
			err = corePtr.RemoveOutbound(tag)
			if err != nil && err != os.ErrInvalid {
				return err
			}
		}
		err = tx.Where("tag = ?", tag).Delete(model.Outbound{}).Error
		if err != nil {
			return err
		}
	default:
		return common.NewErrorf("unknown action: %s", act)
	}
	return nil
}

// ReplaceOutboundItem is one target in an atomic "replace IP" batch. Tag, not
// id, is the stable target: users and routes keep referring to the same tag.
type ReplaceOutboundItem struct {
	Tag        string `json:"tag"`
	Server     string `json:"server"`
	ServerPort uint16 `json:"server_port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

type ReplaceOutboundsResult struct {
	Saved     int                      `json:"saved"`
	Applied   bool                     `json:"applied"`
	Warning   string                   `json:"warning,omitempty"`
	Outbounds []map[string]interface{} `json:"outbounds"`
}

// ReplaceServers updates server/server_port/username/password in place for a
// batch of SOCKS/HTTP outbounds. All rows are validated before any write; a bad
// target rolls the whole batch back. The other options, id, tag and type are
// preserved byte-for-byte at the JSON value level.
func (o *OutboundService) ReplaceServers(items []ReplaceOutboundItem) (*ReplaceOutboundsResult, error) {
	if len(items) == 0 {
		return nil, common.NewError("no outbound targets")
	}
	db := database.GetDB()
	tx := db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	seen := make(map[string]bool, len(items))
	for _, item := range items {
		tag := strings.TrimSpace(item.Tag)
		if tag == "" {
			return nil, common.NewError("outbound target tag is empty")
		}
		if seen[tag] {
			return nil, common.NewErrorf("outbound %q appears more than once", tag)
		}
		seen[tag] = true
		if item.ServerPort == 0 {
			return nil, common.NewErrorf("outbound %q: port must be 1-65535", tag)
		}
		server := strings.TrimSpace(item.Server)
		if server == "" {
			return nil, common.NewErrorf("outbound %q: server is empty", tag)
		}

		var outbound model.Outbound
		if err := tx.Where("tag = ?", tag).First(&outbound).Error; err != nil {
			return nil, common.NewErrorf("outbound %q: %v", tag, err)
		}
		kind := strings.ToLower(outbound.Type)
		if kind != "socks" && kind != "http" {
			return nil, common.NewErrorf("outbound %q: type %q is not supported", tag, outbound.Type)
		}

		options := make(map[string]json.RawMessage)
		if len(outbound.Options) > 0 {
			if err := json.Unmarshal(outbound.Options, &options); err != nil {
				return nil, common.NewErrorf("outbound %q: invalid options: %v", tag, err)
			}
		}
		options["server"] = mustJSON(server)
		options["server_port"] = mustJSON(item.ServerPort)
		// Empty credentials mean "no auth" in the existing SOCKS/HTTP forms.
		setOrDeleteOption(options, "username", item.Username)
		setOrDeleteOption(options, "password", item.Password)
		encoded, err := json.MarshalIndent(options, "", "  ")
		if err != nil {
			return nil, err
		}
		if err = tx.Model(model.Outbound{}).Where("id = ?", outbound.Id).
			Update("options", encoded).Error; err != nil {
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	result := &ReplaceOutboundsResult{Saved: len(items)}
	if corePtr == nil {
		result.Warning = "configuration saved, but the core is not available"
	} else if err := (&ConfigService{}).RestartCore(); err != nil {
		result.Warning = fmt.Sprintf("configuration saved, but applying it failed: %v", err)
	} else {
		result.Applied = true
	}
	if list, err := o.GetAll(); err == nil {
		result.Outbounds = *list
	} else if result.Warning == "" {
		result.Warning = fmt.Sprintf("configuration saved, but reloading the outbound list failed: %v", err)
	}
	return result, nil
}

func mustJSON(value interface{}) json.RawMessage {
	encoded, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`null`)
	}
	return encoded
}

func setOrDeleteOption(options map[string]json.RawMessage, key, value string) {
	if value == "" {
		delete(options, key)
		return
	}
	options[key] = mustJSON(value)
}
