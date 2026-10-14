# Changelog

## 0.9.0

Breaking release. The module path stays `github.com/Decane-io/decane-go` (v0 needs no suffix).

Breaking changes

- Removed `PaymentLinks`, `Subscriptions` and the old deposit-wallet calls (`Wallets.List/Get/Generate/Balance`). They called portal routes that never accepted an API key. Manage them in the portal.
- `Wallets.List` now returns your named wallets with balances (`GET /vendor/query/balances`).
- Swaps use the real contract: `SwapQuoteParams` and `SwapParams` take `FromNetwork`, `FromCurrency`, `FromAmount`, `ToNetwork`, `ToCurrency` (plus `QuoteID`, `SourceWalletID`, `DestinationWalletID` or `DestinationAddress`, `Slippage`). The old `Amount` body returned 422 on every call. Paths are `/vendor/swap/quote`, `/vendor/swap`, `/vendor/swap/{id}`, `/vendor/swaps`. `Swap.ID` reads `swap_id`.
- `Swaps.List` takes `SwapListParams` (limit, offset, status) and returns `SwapList{Swaps, Total}`.
- Lists use cursor pagination: `ListParams{Limit, StartingAfter, Status}` and `List[T]{Data, HasMore, NextCursor}`. `PaginatedList`, `Page`, `Currency` and `Network` filters are gone.
- `Payouts.Get` and `Payouts.List` use `GET /payouts/{id}` and `GET /payouts`.
- `Payouts.CreateBatch` takes `*CreateBatchParams` and returns `*PayoutBatch`.
- Money is a decimal string. `CreateInvoiceParams.Amount`, `CreatePayoutParams.Amount` and the swap amounts are `string`. Response amounts are the `Amount` string type. `FormatAmount(float64)` converts a float without exponent notation.
- `Invoice.WalletAddress` is now `DepositAddress`, `PaidAmount` is `ReceivedAmount`, `Metadata` is `ExtraData`.
- Legacy `X-Granthe-*` and `X-Flow-*` header fallbacks removed. `VerifyRequest` is deprecated.
- Error text starts with `decane:`. Errors carry `Type`, `Param` and `RequestID`.

Added

- `Balances.Get`, `Wallets.List`, `Events.List`, `Events.Resend`, `Invoices.Simulate` (test keys).
- `SourceWalletID` on payouts.
- `PermissionError` (403) and `ConflictError` (409).
- `WebhookEvent.Created`. Event names are `invoice.completed`, not `invoice.paid`; data uses `invoice_id` and `payout_id`.
- `Version` and the User-Agent are `0.9.0`.
