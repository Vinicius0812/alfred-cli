# Alfred Interface Identity

Alfred's interface is inspired by the calm, discreet, and dependable presence
of a trusted technical assistant. Personality should make the CLI recognizable
without making security decisions, errors, or automation harder to understand.

## Voice

Human-facing messages are:

- courteous, concise, and composed;
- direct about risk, failure, and the action the user must take;
- available in English and Brazilian Portuguese;
- addressed to “Master Bruce” or “mestre Bruce” at meaningful moments rather
  than on every line.

Examples include “How may I assist you, Master Bruce?” and “Como posso
ajudá-lo, mestre Bruce?”. Confirmations state the step, risk reason, and safe
default before using the character voice.

## Visual mark

The text bat and `ALFRED` wordmark identify the interactive menu. They use
plain Unicode-independent terminal characters so the menu remains readable on
Windows, Linux, and macOS. The mark is decorative and never replaces a status,
warning, or error label.

## Automation boundary

Personality belongs to human-readable text. It must not alter:

- JSON field names, enum values, timestamps, or exit codes;
- subprocess output;
- audit schema or configuration syntax;
- completion tokens and executable names.

When `--output json` is selected, stdout contains one valid JSON document.
Interactive prompts and subprocess streams use stderr so scripts can parse
stdout safely. JSON values remain neutral and stable across interface
languages.

## Writing new messages

Every new human message should be added to both language catalogs and tested at
the command boundary. Safety information comes first: name the operation,
explain why confirmation is required, show the conservative default, and only
then add personality. Avoid jokes, ambiguous success messages, excessive
honorifics, or references that hide the underlying technical state.

This separation lets Alfred have a distinct identity while remaining an
auditable development tool rather than a conversational agent.
