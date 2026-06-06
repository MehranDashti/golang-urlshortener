<?php


function a(array $orders): array 
{
    $result = [];
    foreach($orders as $order) {
        if ($order['status'] === 'completed') {
            $result[$order['user_id']] = ($result[$order['user_id']] ?? 0) + 1;
        }
    }

    return $result;
}

$transactions = [
    ['user_id' => 1, 'amount' => 150, 'status' => 'completed'],
    ['user_id' => 2, 'amount' => 200, 'status' => 'completed'],
    ['user_id' => 1, 'amount' => 100, 'status' => 'failed'],
    ['user_id' => 3, 'amount' => 300, 'status' => 'completed'],
    ['user_id' => 2, 'amount' => 50,  'status' => 'completed'],
    ['user_id' => 1, 'amount' => 75,  'status' => 'completed'],
];

$result = getOrderReports();

print_r($result);