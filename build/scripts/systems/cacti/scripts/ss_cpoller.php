<?php

error_reporting(0);

if (!isset($called_by_script_server)) {
	include_once(dirname(__FILE__) . '/../include/cli_check.php');

	array_shift($_SERVER['argv']);

	print call_user_func_array('ss_cpoller', $_SERVER['argv']);
}

function ss_cpoller($cmd, $arg1 = '', $arg2 = '') {
	if ($cmd == 'index') {
		$collectors = db_fetch_assoc('SELECT id FROM poller ORDER BY id');

		if (cacti_sizeof($collectors)) {
			foreach ($collectors as $collector) {
				print $collector['id'] . PHP_EOL;
			}
		}
	} elseif ($cmd == 'query') {
		$arg = $arg1;

		if ($arg1 == 'pollerId') {
			$arr = db_fetch_assoc('SELECT id FROM poller ORDER BY id');

			if (cacti_sizeof($arr)) {
				foreach ($arr as $item) {
					print $item['id'] . '!' . $item['id'] . PHP_EOL;
				}
			}
		} elseif ($arg1 == 'pollerName') {
			$arr = db_fetch_assoc('SELECT id, name FROM poller ORDER BY id');

			if (cacti_sizeof($arr)) {
				foreach ($arr as $item) {
					print $item['id'] . '!' . $item['name'] . PHP_EOL;
				}
			}
		}
	} elseif ($cmd == 'get') {
		$arg   = $arg1;
		$index = $arg2;
		$value = '0';

		switch($arg) {
			case 'recacheTime':
				$value = '0';
				$stats = explode(' ', db_fetch_cell('SELECT value FROM settings WHERE name="stats_recache_' . $index . '"'));

				foreach($stats as $_stat) {
					if (preg_match('/^RecacheTime:/', $_stat)) {
						$parts = explode(':', $_stat);
						$value = $parts[1];;
					}
				}

				break;
			case 'recacheDevices':
				$value = '0';
				$stats = explode(' ', db_fetch_cell('SELECT value FROM settings WHERE name="stats_recache_' . $index . '"'));

				foreach($stats as $_stat) {
					if (preg_match('/^DevicesRecached:/', $_stat)) {
						$parts = explode(':', $_stat);
						$value = $parts[1];;
					}
				}

				break;
			case 'avgTime':
				$value = db_fetch_cell_prepared('SELECT max_time
					FROM poller
					WHERE id = ?',
					array($index));

				break;
			case 'minTime':
				$value = db_fetch_cell_prepared('SELECT min_time
					FROM poller
					WHERE id = ?',
					array($index));

				break;
			case 'maxTime':
				$value = db_fetch_cell_prepared('SELECT max_time
					FROM poller
					WHERE id = ?',
					array($index));

				break;
			case 'processCount':
				$value = db_fetch_cell_prepared('SELECT processes
					FROM poller
					WHERE id = ?',
					array($index));

				break;
			case 'threadCount':
				$value = db_fetch_cell_prepared('SELECT threads
					FROM poller
					WHERE id = ?',
					array($index));

				break;
			case 'pollerTime':
				$value = db_fetch_cell_prepared('SELECT total_time
					FROM poller
					WHERE id = ?',
					array($index));

				break;
			case 'getSNMP':
				$value = db_fetch_cell_prepared('SELECT snmp
					FROM poller
					WHERE id = ?',
					array($index));

				break;
			case 'getScript':
				$value = db_fetch_cell_prepared('SELECT script
					FROM poller
					WHERE id = ?',
					array($index));

				break;
			case 'getScriptServer':
				$value = db_fetch_cell_prepared('SELECT server
					FROM poller
					WHERE id = ?',
					array($index));

				break;
		}

		return ($value == '' ? '0' : $value);
	}
}

