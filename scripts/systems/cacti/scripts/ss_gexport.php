<?php

error_reporting(0);

if (!isset($called_by_script_server)) {
	include_once(dirname(__FILE__) . '/../include/cli_check.php');

	array_shift($_SERVER['argv']);

	print call_user_func_array('ss_gexport', $_SERVER['argv']);
}

function ss_gexport($cmd, $arg1 = '', $arg2 = '') {
	if ($cmd == 'index') {
		$exports = db_fetch_assoc('SELECT id FROM graph_exports ORDER BY id');

		if (cacti_sizeof($exports)) {
			foreach ($exports as $export) {
				print $export['id'] . PHP_EOL;
			}
		}
	} elseif ($cmd == 'query') {
		$arg = $arg1;

		if ($arg1 == 'exportId') {
			$arr = db_fetch_assoc('SELECT id FROM graph_exports ORDER BY id');

			if (cacti_sizeof($arr)) {
				foreach ($arr as $item) {
					print $item['id'] . '!' . $item['id'] . PHP_EOL;
				}
			}
		} elseif ($arg1 == 'exportName') {
			$arr = db_fetch_assoc('SELECT id, name FROM graph_exports ORDER BY id');

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
			case 'lastRuntime':
				$value = db_fetch_cell_prepared('SELECT last_runtime
					FROM graph_exports
					WHERE id = ?',
					array($index));

				break;
			case 'totalGraphs':
				$value = db_fetch_cell_prepared('SELECT total_graphs
					FROM graph_exports
					WHERE id = ?',
					array($index));

				break;
		}

		return ($value == '' ? '0' : $value);
	}
}

