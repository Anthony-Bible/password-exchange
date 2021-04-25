<?php

include_once(dirname(__FILE__) . "/database.php"); 
$db=new database();

error_reporting(E_ALL | E_WARNING | E_NOTICE);
ini_set('display_errors', TRUE);
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
function EncryptMessage($messagedata, $encryption_key){
    $ciphering = "AES-128-CTR"; 
    $iv = openssl_random_pseudo_bytes(openssl_cipher_iv_length($ciphering));
    $encryption = openssl_encrypt($messagedata, $ciphering, $encryption_key, $options=0, $iv); 
    return array($iv, $encryption, $encryption_key);
}
function DecryptMessage($messagedata, $encryption_key, $iv){
    $ciphering = "AES-128-CTR"; 
    $encryption = openssl_decrypt($messagedata, $ciphering, $encryption_key, $options=0, $iv); 
    return array($iv, $encryption, $encryption_key);
}
function storeMessageinDatabase($encryptedstring, $firstname, $lastname)
{   
    global $db;
    echo "storing in database";
    $link=$db;
    echo "connected";
    if (!($encryptedstring==''|| $firstname==''||$lastname='')){
        try{
        $sql = "INSERT INTO messages (message, firstname, lastname) VALUES (:encryptedstring, :firstname, :lastname )";
          
          $sqlData  = array(
                    ":encryptedstring"=>$encryptedstring,
                    ":firstname"=>$firstname,
                    ":lastname"=> $lastname,
                    );
           echo "<errors>".
                " No Errors".
                "</errors>";     
          $stmt = $link->prepare($sql);
          $stmt->execute($sqlData);
                }
                catch (Exception $e) {

                    echo "<response>";
                    echo "<message>";
                    // output error message if fails
                    echo $e->getMessage();
                    echo "\n";
                    echo "</message>";
                    echo "</response>";
        }
  }else{
       echo "<errors>".
            "You need to enter something in the content area".
            "</errors>"; 

  }
echo "</response>";

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