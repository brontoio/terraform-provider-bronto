# Statement ID Rules

## Purpose

Ensuring that all log statements contain a `stmt_id` key, either in the log message or as a log attribute. Also all log 
statements and their `stmt_id` value should be reported in a `statementIds.csv` or `statementIds.json` file located at 
the root of the project.

The file name and line number fields of the `statementIds.csv` or `statementIds.json` are optional. When specified, 
they may represent the location of the line where the log statement associated to a statement ID starts or the location 
of the line where the `stmt_id` attribute is set. These may differ as a log statement may span over multiple lines.
