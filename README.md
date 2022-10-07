# Money #
Utility to import bank statements and manage bank transactions to report on personal finances.

| Date | Status |
|------|--------|
|2022-09-26|Imports standardbank CSV statements into mariadb tables for accounts, bank_accounts, statements transactions. Does not preserve order of transactions on the same date, but do not think that is needed. |
|2022-10-07|Limit statements to cover unique dates in a bank account. During import, ignore dates already imported from other statements.|
|2022-10-07|Change 'other' to 'unknown expense' and 'unknown income'|

Next
* report per account transactions
* identify per bank_account range of dates includes and excluded by statements
* review transactions with "unknown_expense|income" account to assign more specific accounts.
  * create the other account if not exist (e.g. groceries, diesel, ...) and add optional transaction notes.
* identify transfers between accounts (same date + amount with different account)
* merge accounts (all transactions dt/ct against X -> Y and delete account X)
* split accounts (all transactions dt/ct against X -> either X/Y/Z/... and add notes why)
* print reports of transactions where other is not defined
* upload scanned documents and link to transactions, e.g. important receipts, with possible multiple documents on one transaction.