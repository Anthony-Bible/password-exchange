<?php

include_once(dirname(__FILE__) . "/database.php"); 


error_reporting(E_ALL | E_WARNING | E_NOTICE);
ini_set('display_errors', TRUE);
$db = new database();
function redirect($place)
{
	header("Location: $place");
	die();
}
function hashPassword($password)
{
	// echo 'current Version is '.phpversion();
	$hashedPassword = password_hash($password, PASSWORD_DEFAULT);
	return $hashedPassword;
}

function RandomToken($length = 32){
    if(!isset($length) || intval($length) <= 8 ){
      $length = 32;
    }
    if (function_exists('random_bytes')) {
        return bin2hex(random_bytes($length));
    }
    if (function_exists('openssl_random_pseudo_bytes')) {
        return bin2hex(openssl_random_pseudo_bytes($length));
    }
}
function EcryptMessage($messagedata, $encryption_key){
    $ciphering = "AES-128-CTR"; 
    $iv = openssl_random_pseudo_bytes(openssl_cipher_iv_length($cypherMethod));
    $encryption = openssl_encrypt($messagedata, $ciphering, 
    $encryption_key, $options=0, $encryption_iv); 
    return $iv, $encyrption;
}

function mailVerify($usertoken,$firstname,$email){
	$secondEmail=urlencode($email);
	$_SESSION['email']=$email;
	$_SESSION['encodedEmail']=$secondEmail;

	$_SESSION['firstName']=$firstname;
	$_SESSION['token']=$usertoken;
	sendEmail();
	// testInclude3();
	// testInclude2();
}

 ?>