# Table width and API error codes

## Scope

Every rendered Ant Design `Table` and `ProTable` in `light-admin-ui/src/pages` allows dragging the right edge of a column header to change that column's width. Width changes last for the mounted page. Existing table sorting, filters, selection, pagination, and search remain available.

The HTTP response envelope retains `code`, `data`, `page`, and `message`. Success remains `"00000"`. Failure codes become stable strings: `A04xx` for request/auth/permission errors, `B1xxx` for domain conflicts, and `C05xx` for server errors. HTTP status remains meaningful. Known sentinel errors receive dedicated domain codes; otherwise the status determines the general code. The UI continues to consume the existing envelope.

## Architecture

A shared table wrapper owns column widths and header resize handles. Each page swaps its table import to the wrapper. No page owns pointer event logic.

The backend error package maps known errors to codes with `errors.Is` semantics. `echox.Response.JSON` resolves status and code in one place, so controller and middleware responses share the same rules. Tests cover wrapped sentinel errors, default status mappings, and success.

## Verification

Run Go response tests and `go test ./...`; run frontend type checking and build if dependencies are available; inspect every table usage and manually verify pointer dragging, sorting, and selection in the browser if the app can be started.
