package poker

const (
	HighHand 	  Bitmap = 1 << 0
	OnePair 	  Bitmap = 1 << 1
	TwoPair 	  Bitmap = 1 << 2
	ThreeOfAKind  Bitmap = 1 << 3
	Straight      Bitmap = 1 << 4
	Flush   	  Bitmap = 1 << 5
	FullHouse 	  Bitmap = 1 << 6
	FourOfAKind   Bitmap = 1 << 7
	StraightFlush Bitmap = 1 << 8
	RoyalFlush 	  Bitmap = 1 << 9
)

type Hand struct {
	Hand Bitmap `json:"hand"`
}

func checkStraight(mask Bitmap) (bool, bool) {
	royalMask := Bitmap(0x1F << 8) // 5 consecutive bits placed at the end of the bitmap
	if (mask & royalMask) == royalMask {
		return true, true
	}

	for shift := 7; shift >= 0; shift-- { // moved the 5 bits placed at the end towards the init of the map. used to compare against all non-royal flushes.
		straightMask := Bitmap(0x1F << shift)
		if (mask & straightMask) == straightMask {
			return true, false
		}
	}

	wheelMask := Bitmap((1 << 12) | 0xF)
	if (mask & wheelMask) == wheelMask {
		return true, false
	}

	return false, false
}

func GetHand(playerCards []Card, tableCards []Card) Hand {
	cards := append(playerCards, tableCards...)

	rankCounts := make(map[Bitmap]int)
	suitCounts := make(map[Bitmap]int)
	suitMasks := make(map[Bitmap]Bitmap)
	allRanks := Bitmap(0)

	for _, card := range cards {
		rankCounts[card.CardType]++
		suitCounts[card.Naipe]++
		suitMasks[card.Naipe] |= card.CardType
		allRanks |= card.CardType
	}

	var flushSuit Bitmap
	hasFlush := false
	for suit, count := range suitCounts {
		if count >= 5 {
			hasFlush = true
			flushSuit = suit
			break
		}
	}

	if hasFlush {
		isStraight, isRoyal := checkStraight(suitMasks[flushSuit])
		if isRoyal {
			return Hand{Hand: RoyalFlush}
		}
		if isStraight {
			return Hand{Hand: StraightFlush}
		}
	}

	fourOfAKindCount := 0
	threeOfAKindCount := 0
	pairCount := 0

	for _, count := range rankCounts {
		switch count {
		case 4:
			fourOfAKindCount++
		case 3:
			threeOfAKindCount++
		case 2:
			pairCount++
		}
	}

	if fourOfAKindCount > 0 {
		return Hand{Hand: FourOfAKind}
	}
	if (threeOfAKindCount >= 1 && pairCount >= 1) || threeOfAKindCount >= 2 {
		return Hand{Hand: FullHouse}
	}
	if hasFlush {
		return Hand{Hand: Flush}
	}

	isStraight, _ := checkStraight(allRanks)
	if isStraight {
		return Hand{Hand: Straight}
	}

	if threeOfAKindCount > 0 {
		return Hand{Hand: ThreeOfAKind}
	}
	if pairCount >= 2 {
		return Hand{Hand: TwoPair}
	}
	if pairCount == 1 {
		return Hand{Hand: OnePair}
	}

	return Hand{Hand: HighHand}
}