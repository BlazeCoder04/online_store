package validate

import (
	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"
)

func ValidateRequest(req proto.Message) error {
	v, err := protovalidate.New()
	if err != nil {
		return err
	}

	if err := v.Validate(req); err != nil {
		return err
	}

	return nil
}
