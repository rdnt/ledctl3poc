package screen

//import (
//	"ledctl3/pkg/screencapture/dxgi"
//	types2 "ledctl3/pkg/screencapture/types"
//	"ledctl3/pkg/uuid"
//	"ledctl3/source/types"
//)
//
//type InputProvider struct {
//	displayRepo types2.DisplayRepository
//	ids         map[int]uuid.UUID
//}
//
//func New() (*InputProvider, error) {
//	dr, err := dxgi.New()
//	if err != nil {
//		return nil, err
//	}
//
//	return &InputProvider{
//		displayRepo: dr,
//		ids:         make(map[int]uuid.UUID),
//	}, nil
//}
//
//func (p *InputProvider) Inputs() ([]*Input, error) {
//	displays, err := p.displayRepo.All()
//	if err != nil {
//		return nil, err
//	}
//
//	var ins []*Input
//	for _, d := range displays {
//		p.ids[d.Id()] = uuid.New()
//
//		ins = append(ins, &Input{
//			id:      p.ids[d.Id()],
//			events:  make(chan types.UpdateEvent),
//			display: d,
//			outputs: nil,
//		})
//	}
//
//	return ins, nil
//}
