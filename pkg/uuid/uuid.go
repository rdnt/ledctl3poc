package uuid3

//
//type UUID uuid.UUID
//
//func New() UUID {
//	return UUID(uuid.New())
//}
//
//var Nil = UUID(uuid.Nil)
//
//func Parse(s string) (UUID, error) {
//	uid, err := uuid.Parse(s)
//	if err != nil {
//		return Nil, err
//	}
//
//	return UUID(uid), nil
//}
//
//func MustParse(s string) UUID {
//	uid, err := uuid.Parse(s)
//	if err != nil {
//		panic(err)
//	}
//
//	return UUID(uid)
//}
