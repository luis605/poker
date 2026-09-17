package poker

type Bitmap uint64

const (
	SuitClubs    Bitmap = 1 << 0 // ♣
	SuitDiamonds Bitmap = 1 << 1 // ♦
	SuitHearts   Bitmap = 1 << 2 // ♥
	SuitSpades   Bitmap = 1 << 3 // ♠
)

const (
	RankTwo   Bitmap = 1 << 0
	RankThree Bitmap = 1 << 1
	RankFour  Bitmap = 1 << 2
	RankFive  Bitmap = 1 << 3
	RankSix   Bitmap = 1 << 4
	RankSeven Bitmap = 1 << 5
	RankEight Bitmap = 1 << 6
	RankNine  Bitmap = 1 << 7
	RankTen   Bitmap = 1 << 8
	RankJack  Bitmap = 1 << 9
	RankQueen Bitmap = 1 << 10
	RankKing  Bitmap = 1 << 11
	RankAce   Bitmap = 1 << 12
)

type Card struct {
	Naipe    Bitmap `json:"naipe"`
	CardType Bitmap `json:"card_type"`
}

func NewCard(rank Bitmap, suit Bitmap) Card {
	return Card{
		CardType: rank,
		Naipe:    suit,
	}
}