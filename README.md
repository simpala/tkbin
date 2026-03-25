# tkbin
token paking with go
tkbin: A Token-Based Object Store
tkbin converts text and code into a binary "pixel" format (4 tokens per 8-byte block). It's designed for LLM agents to perform high-speed searches and metadata-based filtering without loading raw text into memory.

Key Features:

Pixel-Aligned: 8-byte fixed-width records for O(1) seek times.

Metadata-Rich: Tag files with language, category, or AI-generated summaries in a portable JSON index.

Searchable: Direct binary search on token sequences.

Example usage:

cd examples

go run all.go

go run viewer.go list
go run viewer.go info hello.go
go run viewer.go view hello.go
