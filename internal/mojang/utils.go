package mojang

import (
	"fmt"
)

func FormatUUID(uuid string) string {
	if len(uuid) != 32 {
		return uuid
	}

	return fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		uuid[0:8],
		uuid[8:12],
		uuid[12:16],
		uuid[16:20],
		uuid[20:32],
	)
}
