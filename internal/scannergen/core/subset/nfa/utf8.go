package nfa

import (
	"unicode/utf8"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/frontend"
)

const (
	// MaxUTF8Rune1Byte is the maximum Unicode codepoint which can be encoded as 1 byte.
	MaxUTF8Rune1Byte = 1<<7 - 1

	// MaxUTF8Rune2Bytes is the maximum Unicode codepoint which can be encoded as 2 bytes.
	MaxUTF8Rune2Bytes = 1<<11 - 1

	// MaxUTF8Rune3Bytes is the maximum Unicode codepoint which can be encoded as 3 bytes.
	MaxUTF8Rune3Bytes = 1<<16 - 1

	// UTF8LeadingByte2Prefix is the prefix OR'd into the leading byte of a 2-byte sequence (110xxxxx).
	UTF8LeadingByte2Prefix rune = 0xC0

	// UTF8LeadingByte3Prefix is the prefix OR'd into the leading byte of a 3-byte sequence (1110xxxx).
	UTF8LeadingByte3Prefix rune = 0xE0

	// UTF8LeadingByte4Prefix is the prefix OR'd into the leading byte of a 4-byte sequence (11110xxx).
	UTF8LeadingByte4Prefix rune = 0xF0

	// UTF8ContinuationBytePrefix is the prefix OR'd into every continuation byte (10xxxxxx).
	UTF8ContinuationBytePrefix rune = 0x80
)

// BuildUTF8Encoding constructs NFA states and transitions which match the given char range when encoded as UTF-8.
// While the char range can be within 0x00 to unicode.MaxRune, the emitted NFA states and transitions are always within
// 0x00 to 0xFF.
func BuildUTF8Encoding(charRange frontend.CharRange, ruleIdx int, states []State, startStateIdx int) []State {
	if utf8.RuneLen(charRange.Low) != utf8.RuneLen(charRange.High) {
		panic("char range low and high must encode to the same number of UTF-8 bytes")
	}

	switch {
	case charRange.High <= MaxUTF8Rune1Byte:
		// 1-byte (0xxxxxxx): the rune value is its own byte value.
		states[startStateIdx].Transitions = append(states[startStateIdx].Transitions, Transition{
			ByteRange: backend.ByteRange{
				Low:  byte(charRange.Low),  //nolint:gosec // Cannot overflow because of range check.
				High: byte(charRange.High), //nolint:gosec // Cannot overflow because of range check.
			},
			NextStateIdx: -1,
		})

	case charRange.High <= MaxUTF8Rune2Bytes:
		// 2-byte (110xxxxx 10xxxxxx): leading byte carries the top 5 bits (shifted by 6).
		states = buildUTF8EncodingByte(
			charRange,
			UTF8LeadingByte2Prefix,
			1,
			ruleIdx,
			states,
			startStateIdx,
		)

	case charRange.High <= MaxUTF8Rune3Bytes:
		// 3-byte (1110xxxx 10xxxxxx 10xxxxxx): leading byte carries the top 4 bits (shifted by 12).
		states = buildUTF8EncodingByte(
			charRange,
			UTF8LeadingByte3Prefix,
			2,
			ruleIdx,
			states,
			startStateIdx,
		)

	default:
		// 4-byte (11110xxx 10xxxxxx 10xxxxxx 10xxxxxx): leading byte carries the top 3 bits (shifted by 18).
		states = buildUTF8EncodingByte(
			charRange,
			UTF8LeadingByte4Prefix,
			3,
			ruleIdx,
			states,
			startStateIdx,
		)
	}
	return states
}

// buildUTF8EncodingByte encodes one byte of a multi-byte UTF-8 sequence and recurses for the remaining bytes.
// prefix is OR'd with the extracted payload bits to form the actual byte value.
// remainingBytes is the number of continuation bytes that follow this byte (0 means this is the last byte).
//
//nolint:funlen // Yes, this function is long but not easy to shorten.
func buildUTF8EncodingByte(
	charRange frontend.CharRange,
	prefix rune,
	remainingBytes int,
	ruleIdx int,
	states []State,
	startStateIdx int,
) []State {
	shift := remainingBytes * 6
	continuationMask := rune((1 << shift) - 1)

	thisRange := backend.ByteRange{
		Low:  byte(prefix | (charRange.Low >> shift)),  //nolint:gosec // Cannot overflow because of shift.
		High: byte(prefix | (charRange.High >> shift)), //nolint:gosec // Cannot overflow because of shift.
	}

	if remainingBytes == 0 {
		// Last byte: emit a direct range transition.
		states[startStateIdx].Transitions = append(states[startStateIdx].Transitions, Transition{
			ByteRange:    thisRange,
			NextStateIdx: -1,
		})
		return states
	}

	if thisRange.Low == thisRange.High {
		// Same byte at this position: one intermediate state, then recurse.
		intermediateStateIdx := len(states)
		states = append(states, State{
			RuleIdx: ruleIdx,
		})
		states[startStateIdx].Transitions = append(states[startStateIdx].Transitions, Transition{
			ByteRange:    thisRange,
			NextStateIdx: intermediateStateIdx,
		})
		remainingCharRange := frontend.CharRange{
			Low:  charRange.Low & continuationMask,
			High: charRange.High & continuationMask,
		}
		return buildUTF8EncodingByte(
			remainingCharRange,
			UTF8ContinuationBytePrefix,
			remainingBytes-1,
			ruleIdx,
			states,
			intermediateStateIdx,
		)
	}

	// Byte at this position differs: split into low, middle, and high.
	lowRemainingCharRange := frontend.CharRange{
		Low:  charRange.Low & continuationMask,
		High: continuationMask,
	}
	fullRemainingCharRange := frontend.CharRange{
		Low:  0,
		High: continuationMask,
	}
	highRemainingCharRange := frontend.CharRange{
		Low:  0,
		High: charRange.High & continuationMask,
	}

	// Low: runs from lows payload bits up to the maximum
	{
		intermediateStateIdx := len(states)
		states = append(states, State{
			RuleIdx: ruleIdx,
		})
		states[startStateIdx].Transitions = append(states[startStateIdx].Transitions, Transition{
			ByteRange: backend.ByteRange{
				Low:  thisRange.Low,
				High: thisRange.Low,
			},
			NextStateIdx: intermediateStateIdx,
		})
		states = buildUTF8EncodingByte(
			lowRemainingCharRange,
			UTF8ContinuationBytePrefix,
			remainingBytes-1,
			ruleIdx,
			states,
			intermediateStateIdx,
		)
	}

	// Middle: runs from low to high with full range
	if thisRange.Low+1 <= thisRange.High-1 {
		intermediateStateIdx := len(states)
		states = append(states, State{
			RuleIdx: ruleIdx,
		})
		states[startStateIdx].Transitions = append(states[startStateIdx].Transitions, Transition{
			ByteRange: backend.ByteRange{
				Low:  thisRange.Low + 1,
				High: thisRange.High - 1,
			},
			NextStateIdx: intermediateStateIdx,
		})
		states = buildUTF8EncodingByte(
			fullRemainingCharRange,
			UTF8ContinuationBytePrefix,
			remainingBytes-1,
			ruleIdx,
			states,
			intermediateStateIdx,
		)
	}

	// High: runs from zero up to high
	{
		intermediateStateIdx := len(states)
		states = append(states, State{
			RuleIdx: ruleIdx,
		})
		states[startStateIdx].Transitions = append(states[startStateIdx].Transitions, Transition{
			ByteRange: backend.ByteRange{
				Low:  thisRange.High,
				High: thisRange.High,
			},
			NextStateIdx: intermediateStateIdx,
		})
		states = buildUTF8EncodingByte(
			highRemainingCharRange,
			UTF8ContinuationBytePrefix,
			remainingBytes-1,
			ruleIdx,
			states,
			intermediateStateIdx,
		)
	}

	return states
}
