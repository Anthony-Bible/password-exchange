console.log("pre-submit");
$(document).ready(function(){
$("#contact-form").submit(function( event ){ 
    //Prevent the default action from happening
    console.log("submitted");
    event.preventDefault();
    // grecaptcha.ready(function() {
    //     grecaptcha.execute('6LeFdokUAAAAAHjyx-b7eiMeBrZs4FGgZlyurW9V', {action: ''});
    // });
  
    var Loginform = $( "#contact-form" );

   


 
  try 
     {
        
    var email= Loginform.find( "input[name='email']" ).val();
    var name= Loginform.find( "input[name='name']" ).val();

    var other_lastname= Loginform.find( "input[name='other_lastname']" ).val();
    var other_firstname= Loginform.find( "input[name='other_firstname']" ).val();
    var message = Loginform.find( "textarea[name='message']" ).val();
    var other_email = Loginform.find("input[name='other_email']").val();

    // console.log(token);
    url = Loginform.attr( "action" );

        // console.log(grecaptcha.getResponse());
    posting = $.post( url, {email:email,name:name,other_firstname:other_firstname,other_lastname:other_lastname,other_email:other_email,message:message,captcha:grecaptcha.getResponse()}, function(data, status){
        var xmlDoc = $.parseXML( data ); 
        var $xml = $(xmlDoc);
        var  $person = $xml.find("response");
        $messageDiv = $("#wasitasuccess");
        $person.each(function(){
            console.log("This is the checked");
            
            $checkedResponse="You Successfully sent the information if they don't recieve an email please have them check their spam folder";
            console.log($checkedResponse);
            var $actualResponse = $(this).find('message').text();
            $messageDiv.html($actualResponse);      
            if($actualResponse==$checkedResponse)
            {
            $messageDiv.css("color","green");
            $("#sendinfo").attr("disabled","disabled");
            }else{
                grecaptcha.reset();
                $messageDiv.css("color","red");

            }
            

        });
    });
   }
     catch(e) 
     { 
     $("#contactForm").html(e); 
     }
});});