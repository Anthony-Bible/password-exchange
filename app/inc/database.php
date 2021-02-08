<?php echo '<?xml version="1.0" encoding="UTF-8" ?>'; ?>
<!DOCTYPE math
    PUBLIC "-//W3C//DTD MathML 2.0//EN"
           "http://www.w3.org/Math/DTD/mathml2/mathml2.dtd" >
<?php 

	require(dirname(__FILE__) . "/../vendor/autoload.php");
     	$dotenv = Dotenv\Dotenv::createImmutable(__DIR__ . '/../');
	$dotenv->load();   

echo "<response>";
#print_r($_ENV);
if(session_status() == 1){
	session_start();
}




class database{
 
  var $conn;
    //connect to the database
  function connect() {
      try {
        $user=getenv('PASSWORDEXCHANGEUSER');
        $pass = getenv('PASSWORDEXCHANGEPASS');
        $dbname=getenv('PASSWORDEXCHANGEDBNAME');
        $servername=getenv('PASSWORDEXCHANGEHOST');
    //connect ot database using PDO
    $this->conn = new PDO('mysql:host='.$servername.';dbname='.$dbname, $user, $pass,
          [
                PDO::ATTR_PERSISTENT            => true,
                PDO::ATTR_ERRMODE               => PDO::ERRMODE_EXCEPTION
            ]);
    //$this->conn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    
    //use prepared statements for user submitted information
     return $this->conn;
	
    
    } catch (PDOException $e) {
      echo 'ERROR: '. $e->getMessage();
      echo "<br><br>";
      print_r($e);
    
    }
  }

  //this will close the database connection
  function close() {
    $this->conn = null;
  }
  
    //get the data filled out of the registration and put it in the database
  function createUser(array $values)
  {
    $link=$this->connect();    	
       
        //pass $_POST in and use that to get the values from the form
  	$first_name=$values['firstName'];
    $last_name=$values['lastName'];
    $username=$values['username'];
  	$password=$values['password'];
    $email =$values['email'];
    $dob =$values['date'];
    $phone =$values['phone'];
    	
       
    //get todays date for created date
    //will have the time set local to the server for now but will change for user's timezone
    date_default_timezone_set("America/Denver");
          
    $password=hashPassword($password);

    // Check the values to make sure there won't be duplicates
		if ((checkUsername($username, $this) || checkEmail($email, $this) || checkPhone($phone, $this))) {
			# code...

		}else{
			$sql = "INSERT INTO users (first_name, last_name, dateOfBirth, email,username, password, phone, created,modified) VALUES (:fname, :lname, :dob,:email, :username, :pass,:phone, now(),now() )";
    		$sqlData  = array(
    				  ":fname"=>$first_name,
                      ":lname"=>$last_name,
                      ":dob"=>$dob, 
                      ":email"=>$email,
                      ":username" =>$username,
                      ":pass"=> $password,
                      ":phone" =>$phone
                       );
    		 echo "<errors>".
				  " No Errors".
                  "</errors>";     
		    $stmt = $link->prepare($sql);
        $stmt->execute($sqlData);
        $userToken = $this->createToken($email,$username);
        mailVerify($userToken,$first_name,$email);
        
        // $_SESSION['user_id']=$this->getuserID($username);
            // $_SESSION['user_agent']=$_SERVER['HTTP_USER_AGENT'];

		}
    
    // redirect('wall.php');
          echo "</response>";

		
		$this->close();
  }
  


  
    //check username and password against those in the database
function createToken($email,$username){
  $link=$this->connect();
  $currentTime=time();
  $token=md5($email.$currentTime);
  $userID=$this->getuserID($username);

  $mysqlTimeStamp=date('Y-m-d H:i:s', $currentTime);

  $sql = "INSERT INTO verification (user_id,token,timeCreated,verified) VALUES (:userId, :token, :timeCreated, 0)";
  $sqlData  = array(
        ":userId"=>$userID,
                ":token"=>$token,
                ":timeCreated"=>$mysqlTimeStamp
                 );
        
  $stmt = $link->prepare($sql);
  $stmt->execute($sqlData);
  return $token;


}

}



 ?>

