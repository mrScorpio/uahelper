# Data archiver with trends

This tool helps to store data from AstraRegul OPC UA server and shows stored data with echarts.
List of OPC UA tags must be located in file named "tags".
By default storage is the zipped files created once an hour.
Tool can send archives and some messages to telegram bot.
Config must be read from .env file (see .env_example).

TODO:
- add modbus connection
- add subscribtion mode to opc ua