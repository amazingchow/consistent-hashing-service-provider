package common

type Uint8_n []uint8 // nolint

func (e Uint8_n) Len() int { return len(e) }

func (e Uint8_n) Less(i, j int) bool { return e[i] < e[j] }

func (e Uint8_n) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

type Uint16_n []uint16 // nolint

func (e Uint16_n) Len() int { return len(e) }

func (e Uint16_n) Less(i, j int) bool { return e[i] < e[j] }

func (e Uint16_n) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

type Uint32_n []uint32

func (e Uint32_n) Len() int { return len(e) }

func (e Uint32_n) Less(i, j int) bool { return e[i] < e[j] }

func (e Uint32_n) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

type Uint64s_n []uint64 // nolint

func (e Uint64s_n) Len() int { return len(e) }

func (e Uint64s_n) Less(i, j int) bool { return e[i] < e[j] }

func (e Uint64s_n) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
