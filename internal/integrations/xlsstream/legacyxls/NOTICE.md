This directory contains the BIFF8 `.xls` reader subset from
`github.com/jloor/go-excel-reader` v0.3.0, source commit
`4d7a9f10d6d72c5cdbfb57ece9108f703da03579`, under the MIT License in `LICENSE`.

Majucau applies local safety limits before row allocation: no more than 256
columns per row, and shared-string capacity is bounded by the available input
bytes. The upstream reader is otherwise retained with its source comments.
